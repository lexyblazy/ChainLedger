"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";

import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

import { useInfiniteQuery, InfiniteData } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { Suspense, useMemo } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { makeRequest } from "@/lib/api";
import { BalanceSnapshotsResponse } from "@/lib/types";
import { formatTokenBalance } from "@/lib/format";
import { LoadingState } from "@/components/loading-state";

function SnapshotPageInner() {
  const searchParams = useSearchParams();

  const address = searchParams.get("address");
  const chainId = searchParams.get("chainId");
  const tokenAddress = searchParams.get("tokenAddress") ?? "native";

  const enabled = Boolean(address && chainId);

  /* ------------------------------------------------------------------ */
  /*  Query                                                             */
  /* ------------------------------------------------------------------ */

  const {
    data,
    isLoading,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery<
    BalanceSnapshotsResponse,
    Error,
    InfiniteData<BalanceSnapshotsResponse>,
    [string, string | null, string | null, string],
    number | undefined
  >({
    queryKey: ["snapshots", address, chainId, tokenAddress],
    initialPageParam: undefined,
    enabled,
    queryFn: async ({ pageParam }) => {
      const cursorQuery = pageParam ? `&cursor_id=${pageParam}` : "";

      return makeRequest<BalanceSnapshotsResponse>(
        `/wallets/${address}/balance-snapshots?chain_id=${chainId}&token_address=${tokenAddress}${cursorQuery}&limit=500`,
      );
    },
    getNextPageParam: (lastPage) =>
      lastPage.meta.has_more ? lastPage.meta.cursor_id : undefined,
  });

  /* ------------------------------------------------------------------ */
  /*  Token Metadata (must be derived before normalization)             */
  /* ------------------------------------------------------------------ */

  const tokenMetadata = useMemo(() => {
    if (!data?.pages?.length) return null;
    return data.pages[0].data.token_metadata ?? null;
  }, [data]);

  const symbol = tokenMetadata?.symbol ?? "TOKEN";

  const decimals =
    typeof tokenMetadata?.decimals === "number" &&
    Number.isFinite(tokenMetadata.decimals)
      ? tokenMetadata.decimals
      : 18;

  /* ------------------------------------------------------------------ */
  /*  Flatten + Normalize                                               */
  /* ------------------------------------------------------------------ */

  const normalizedSnapshots = useMemo(() => {
    if (!data) return [];

    const flat = data.pages.flatMap((page) => page.data.balance_snapshots);

    return flat.map((s) => {
      const raw = BigInt(s.balance_raw);

      // Convert safely AFTER decimal scaling
      const numeric = Number(raw) / 10 ** decimals;

      return {
        timestamp: new Date(s.block_timestamp).getTime(),
        raw: s.balance_raw,
        value: numeric,
      };
    });
  }, [data, decimals]);

  /* ------------------------------------------------------------------ */
  /*  Autoscale Logic                                                   */
  /* ------------------------------------------------------------------ */

  const scaleInfo = useMemo(() => {
    if (normalizedSnapshots.length === 0) {
      return { scale: 1, short: "", long: "" };
    }

    let max = 0;

    for (const d of normalizedSnapshots) {
      const abs = Math.abs(d.value);
      if (abs > max) max = abs;
    }

    const scales = [
      { scale: 1e18, short: "E", long: "Quintillions" },
      { scale: 1e15, short: "P", long: "Quadrillions" },
      { scale: 1e12, short: "T", long: "Trillions" },
      { scale: 1e9, short: "B", long: "Billions" },
      { scale: 1e6, short: "M", long: "Millions" },
      { scale: 1e3, short: "K", long: "Thousands" },
    ];

    for (const s of scales) {
      if (max >= s.scale) {
        return s;
      }
    }

    return { scale: 1, short: "", long: "" };
  }, [normalizedSnapshots]);

  /* ------------------------------------------------------------------ */
  /*  Chart Data                                                        */
  /* ------------------------------------------------------------------ */

  const chartData = useMemo(() => {
    return normalizedSnapshots
      .map((d) => ({
        timestamp: d.timestamp,
        balance: d.value / scaleInfo.scale,
        raw: d.raw,
      }))
      .sort((a, b) => a.timestamp - b.timestamp);
  }, [normalizedSnapshots, scaleInfo.scale]);

  const latestBalance =
    chartData.length > 0 ? chartData[chartData.length - 1].balance : 0;

  /* ------------------------------------------------------------------ */
  /*  Early Returns                                                     */
  /* ------------------------------------------------------------------ */

  if (!enabled) return <div>Select a wallet.</div>;
  if (isLoading) return <div>Loading snapshots...</div>;
  if (error) return <div>Failed to load snapshots.</div>;

  /* ------------------------------------------------------------------ */
  /*  Render                                                            */
  /* ------------------------------------------------------------------ */

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Back Link */}
      <Link
        href={`/dashboard/portfolio?address=${address}&chainId=${chainId}`}
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Portfolio
      </Link>

      {/* Current Balance */}
      <div>
        <div className="text-sm text-muted-foreground">Current Balance</div>

        <div className="flex items-baseline gap-3">
          <div className="text-3xl font-mono font-semibold">
            {latestBalance.toLocaleString(undefined, {
              minimumFractionDigits: 2,
              maximumFractionDigits: 4,
            })}{" "}
            {symbol}
          </div>

          {scaleInfo.long && (
            <Badge variant="secondary" className="font-mono">
              {scaleInfo.long}
            </Badge>
          )}
        </div>
      </div>

      {/* Chart Card */}
      <Card>
        <CardHeader>
          <CardTitle>
            {symbol} Balance History
            {scaleInfo.long && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                ({scaleInfo.long})
              </span>
            )}
          </CardTitle>

          <CardDescription>Historical balance snapshots</CardDescription>
        </CardHeader>

        <Separator />

        <CardContent className="pt-6">
          <ChartContainer
            className="h-[420px]"
            config={{
              balance: {
                label: "Balance",
                color: "hsl(var(--primary))",
              },
            }}
          >
            <AreaChart
              data={chartData}
              margin={{ top: 20, right: 20, left: 10, bottom: 30 }}
            >
              <defs>
                <linearGradient
                  id="balanceGradient"
                  x1="0"
                  y1="0"
                  x2="0"
                  y2="1"
                >
                  <stop
                    offset="5%"
                    stopColor="var(--color-balance)"
                    stopOpacity={0.4}
                  />
                  <stop
                    offset="95%"
                    stopColor="var(--color-balance)"
                    stopOpacity={0}
                  />
                </linearGradient>
              </defs>

              <CartesianGrid strokeDasharray="3 3" vertical={false} />

              <XAxis
                dataKey="timestamp"
                type="number"
                scale="time"
                domain={["dataMin", "dataMax"]}
                tickFormatter={(value) => new Date(value).toLocaleDateString()}
                tickMargin={12}
              />

              <YAxis
                tickFormatter={(value) =>
                  `${value.toLocaleString(undefined, {
                    maximumFractionDigits: 2,
                  })} ${scaleInfo.short}`
                }
                width={80}
              />

              <ChartTooltip
                content={
                  <ChartTooltipContent
                    indicator="line"
                    formatter={(_, __, props: any) => {
                      const raw = props?.payload?.raw;
                      if (!raw) return "-";
                      return `${formatTokenBalance(raw, decimals)} ${symbol}`;
                    }}
                  />
                }
              />

              <Area
                type="monotone"
                dataKey="balance"
                stroke="var(--color-balance)"
                fill="url(#balanceGradient)"
                strokeWidth={2}
                dot={false}
                isAnimationActive
                animationDuration={600}
              />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>

      {hasNextPage && (
        <div className="flex justify-center">
          <Button
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
            variant="secondary"
          >
            {isFetchingNextPage ? "Loading..." : "Load Older Snapshots"}
          </Button>
        </div>
      )}
    </div>
  );
}

export default function SnapshotPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <SnapshotPageInner />
    </Suspense>
  );
}
