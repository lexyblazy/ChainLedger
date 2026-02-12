"use client";

import { Card, CardHeader, CardContent } from "@/components/ui/card";

import { Table, TableBody, TableRow, TableCell } from "@/components/ui/table";

import { Badge } from "@/components/ui/badge";

import { makeRequest } from "@/lib/api";
import { StatusResponse } from "@/lib/types";
import { useQuery } from "@tanstack/react-query";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";

export default function Status() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["status"],
    queryFn: () => makeRequest<StatusResponse>("/status"),
  });

  if (isLoading) {
    return <LoadingState />;
  }

  if (error || !data) {
    return (
      <ErrorState
        error={error || new Error("Unknown error")}
        title="Failed to load system status."
      />
    );
  }

  const networks = Object.values(data.data.networks || {}).sort(
    (a, b) => a.chainId - b.chainId,
  );

  return (
    <div className="space-y-12 max-w-4xl mx-auto">
      {networks.map((network) => (
        <Card key={network.chainId} className="size-full">
          <CardHeader className="flex items-center justify-between">
            <div className="font-medium">{network.name}</div>
            <Badge variant="secondary">{network.chainId}</Badge>
          </CardHeader>

          <CardContent className="pt-2 pb-4">
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell>Best RPC Block</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.bestRpcBlockNumber.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Max Ingested Block</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.ingestedBlocksMaxBlockNumber.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Ingestion Lag (blocks)</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.ingestionLagBlocksCount.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Ingestion Progress (%)</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.ingestionProgressPct.toFixed(2)}%
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Ingested Blocks Count</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.ingestedBlocksCount.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Start Block</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.startBlock.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Tokens Count</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.tokensCount.toLocaleString()}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>Addresses Count</TableCell>
                  <TableCell className="text-right font-mono">
                    {network.addressesCount.toLocaleString()}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
