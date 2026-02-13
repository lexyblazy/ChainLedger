"use client";

import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { makeRequest } from "@/lib/api";
import { LoadingState } from "@/components/loading-state";
import { ErrorState } from "@/components/error-state";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { formatTokenBalance } from "@/lib/format";
import { Portfolio, PortfolioResponse } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { CHAINS } from "@/lib/constants";
import { PaginationControls } from "@/components/pagination-controls";
import { Suspense, useMemo, useState } from "react";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Separator } from "@/components/ui/separator";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

function PortfolioSummary({ assets }: { assets: Portfolio[] }) {
  const nativeAsset = assets.find((a) => a.asset_type === "native");

  const nonZeroAssets = assets.filter(
    (a) => BigInt(a.balance_raw) !== BigInt(0),
  );

  const erc20Count = nonZeroAssets.filter(
    (a) => a.asset_type === "erc20",
  ).length;

  const negativeCount = nonZeroAssets.filter(
    (a) => BigInt(a.balance_raw) < BigInt(0),
  ).length;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card>
        <CardContent className="py-3">
          <div className="text-xs text-muted-foreground">Native Balance</div>
          <div className="font-mono text-lg font-semibold">
            {nativeAsset
              ? formatTokenBalance(
                  nativeAsset.balance_raw,
                  nativeAsset.decimals,
                )
              : "0.0000"}{" "}
            {nativeAsset?.symbol}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="py-4">
          <div className="text-xs text-muted-foreground">Assets (Non-Zero)</div>
          <div className="text-lg font-semibold">{nonZeroAssets.length}</div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="py-4">
          <div className="text-xs text-muted-foreground">ERC20 Positions</div>
          <div className="text-lg font-semibold">{erc20Count}</div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="py-4">
          <div className="text-xs text-muted-foreground">
            Negative Positions
          </div>
          <div className="text-lg font-semibold text-destructive">
            {negativeCount}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function PortfolioPageInner() {
  const searchParams = useSearchParams();

  const address = searchParams.get("address");
  const chainId = searchParams.get("chainId");

  const enabled = Boolean(address && chainId);
  const [offset, setOffset] = useState(0);
  const [hideZero, setHideZero] = useState(true);
  const limit = 50;

  const { data, isLoading, error } = useQuery({
    queryKey: ["portfolio", address, chainId, offset, limit],
    queryFn: () =>
      makeRequest<PortfolioResponse>(
        `/wallets/${address}/portfolio?chain_id=${chainId}&limit=${limit}&offset=${offset}`,
      ),
    enabled,
  });

  const { total } = data?.meta ?? { total: 0 };
  const assets = data?.data.portfolio ?? [];

  const visibleAssets = hideZero
    ? assets.filter((a) => BigInt(a.balance_raw) !== BigInt(0))
    : assets;

  const sortedPortfolio = useMemo(() => {
    return [...visibleAssets].sort((a, b) => {
      const aVal = BigInt(a.balance_raw);
      const bVal = BigInt(b.balance_raw);

      if (a.asset_type === "native") return -1;
      if (b.asset_type === "native") return 1;

      if (bVal > aVal) return 1;
      if (bVal < aVal) return -1;
      return 0;
    });
  }, [visibleAssets]);

  if (!address || !chainId) {
    return <div>Select a wallet.</div>;
  }

  if (isLoading) return <LoadingState />;

  if (error || !data) {
    return (
      <ErrorState
        error={error || new Error("Unknown error")}
        title="Failed to load portfolio."
      />
    );
  }

  const chainMeta = CHAINS[Number(chainId) as keyof typeof CHAINS];

  return (
    <div className="space-y-4 max-w-6xl mx-auto">
      <Link
        href={`/dashboard/wallets`}
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Wallets
      </Link>
      <div className="flex justify-between">
        <div className="font-mono text-sm">Wallet: {address}</div>
        <Badge variant="secondary">{chainMeta?.name}</Badge>
      </div>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Checkbox
            id="hide-zero"
            checked={hideZero}
            onCheckedChange={(value) => {
              setHideZero(Boolean(value));
              setOffset(0);
            }}
          />
          <Label htmlFor="hide-zero" className="text-sm text-muted-foreground">
            Hide zero balances
          </Label>
        </div>
      </div>

      {chainMeta?.ingestionStartDate && (
        <Card className="border-muted bg-muted/40 shadow-none">
          <CardContent className="text-sm text-muted-foreground">
            This portfolio view reflects balances indexed starting{" "}
            {chainMeta.ingestionStartDate}. Historical balances prior to that
            date are not included. As a result, some assets may appear negative
            due to partial ingestion of transfer history.
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent>
          <PortfolioSummary assets={sortedPortfolio} />
          <Separator className="my-4" />
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Asset</TableHead>
                <TableHead>Type</TableHead>
                <TableHead className="text-right">Balance</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {sortedPortfolio.map((asset) => {
                const displaySymbol =
                  asset.symbol.length > 32
                    ? asset.symbol.slice(0, 32) + "..."
                    : asset.symbol;

                return (
                  <TableRow key={asset.asset_address}>
                    <TableCell>
                      <Link
                        href={{
                          pathname: "/dashboard/snapshots",
                          query: {
                            address,
                            chainId,
                            tokenAddress: asset.asset_address,
                          },
                        }}
                        className="block hover:opacity-80 transition"
                      >
                        <div className="font-semibold text-sm hover:underline">
                          {displaySymbol}
                        </div>

                        {asset.asset_type !== "native" && (
                          <div className="text-xs text-muted-foreground font-mono">
                            {asset.asset_address}
                          </div>
                        )}
                      </Link>
                    </TableCell>

                    <TableCell className="text-xs text-muted-foreground">
                      {asset.asset_type}
                    </TableCell>

                    <TableCell
                      className={`text-right font-mono ${
                        asset.balance_raw.startsWith("-")
                          ? "text-destructive"
                          : ""
                      }`}
                    >
                      {formatTokenBalance(asset.balance_raw, asset.decimals)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
          <PaginationControls
            offset={offset}
            limit={limit}
            total={total}
            currentCount={sortedPortfolio.length}
            onPrevious={() =>
              setOffset((prev: number) => Math.max(0, prev - limit))
            }
            onNext={() => setOffset((prev: number) => prev + limit)}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export default function PortfolioPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <PortfolioPageInner />
    </Suspense>
  );
}
