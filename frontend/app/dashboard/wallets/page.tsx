"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { makeRequest } from "@/lib/api";
import { WalletsResponse } from "@/lib/types";
import { CHAINS } from "@/lib/constants";
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
import { ChainSelector } from "@/components/chain-selector";
import { WalletUpsertDialog } from "@/components/wallet-upsert-dialog";
import { useRouter } from "next/navigation";
import { PaginationControls } from "@/components/pagination-controls";
import { formatDate } from "@/lib/format";

const chains = Object.values(CHAINS);

export default function WalletsPage() {
  const defaultChainId = chains[0]?.chainId ?? 1;
  const [chainId, setChainId] = useState(defaultChainId);
  const [offset, setOffset] = useState(0);
  const limit = 100;
  const router = useRouter();

  const { data, isLoading, error } = useQuery({
    queryKey: ["wallets", chainId, offset, limit],
    queryFn: () =>
      makeRequest<WalletsResponse>(
        `/wallets?chain_id=${chainId}&limit=${limit}&offset=${offset}`,
      ),
  });

  const handlePrevious = () => {
    setOffset((prev) => Math.max(0, prev - limit));
  };

  const handleNext = () => {
    setOffset((prev) => prev + limit);
  };

  const handleChainChange = (newChainId: number) => {
    setChainId(newChainId);
    setOffset(0);
  };

  if (isLoading) return <LoadingState />;

  if (error || !data) {
    return (
      <ErrorState
        error={error || new Error("Unknown error")}
        title="Failed to load wallets."
      />
    );
  }

  const { wallets } = data.data;
  const { total } = data.meta;

  return (
    <div className="space-y-4 max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <ChainSelector
          chains={chains}
          value={chainId}
          onChange={handleChainChange}
        />

        <WalletUpsertDialog chainId={chainId} />
      </div>

      <Card>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Address</TableHead>
                <TableHead>Label</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Chain</TableHead>
                <TableHead className="text-right">Updated</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {wallets.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground py-6"
                  >
                    No wallets found for this chain.
                  </TableCell>
                </TableRow>
              )}
              {wallets.map((wallet) => (
                <TableRow
                  key={wallet.address}
                  onClick={() =>
                    router.push(
                      `/dashboard/portfolio?address=${wallet.address}&chainId=${wallet.chain_id}`,
                    )
                  }
                >
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {wallet.address}
                  </TableCell>
                  <TableCell className="py-2 text-xs">{wallet.label}</TableCell>
                  <TableCell className="py-2 text-xs text-muted-foreground">
                    {wallet.entity_type}
                  </TableCell>
                  <TableCell className="py-2 text-xs">
                    {CHAINS[wallet.chain_id as keyof typeof CHAINS]?.name}
                  </TableCell>
                  <TableCell className="text-right py-2 font-mono text-xs text-muted-foreground">
                    {formatDate(wallet.updated_at)}
                  </TableCell>
                  <TableCell
                    className="text-right"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <WalletUpsertDialog chainId={chainId} wallet={wallet} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <PaginationControls
            offset={offset}
            limit={limit}
            total={total}
            currentCount={wallets.length}
            onPrevious={handlePrevious}
            onNext={handleNext}
          />
        </CardContent>
      </Card>
    </div>
  );
}
