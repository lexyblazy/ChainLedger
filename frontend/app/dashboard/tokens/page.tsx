"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { makeRequest } from "@/lib/api";
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
import { PaginationControls } from "@/components/pagination-controls";
import { Token, TokensResponse } from "@/lib/types";
import { Badge } from "@/components/ui/badge";

export default function TokensPage() {
  const chains = Object.values(CHAINS);
  const defaultChainId = chains[0]?.chainId ?? 1;

  const [chainId, setChainId] = useState(defaultChainId);
  const [offset, setOffset] = useState(0);
  const limit = 50;

  const { data, isLoading, error } = useQuery({
    queryKey: ["tokens", chainId, offset, limit],
    queryFn: () =>
      makeRequest<TokensResponse>(
        `/tokens?chain_id=${chainId}&limit=${limit}&offset=${offset}`,
      ),
  });

  if (isLoading) return <LoadingState />;

  if (error || !data) {
    return (
      <ErrorState
        error={error || new Error("Unknown error")}
        title="Failed to load tokens."
      />
    );
  }

  const { tokens } = data.data;
  const { total } = data.meta;

  return (
    <div className="space-y-4 max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <ChainSelector
          chains={chains}
          value={chainId}
          onChange={(newChainId) => {
            setChainId(newChainId);
            setOffset(0);
          }}
        />
      </div>

      <div className="flex justify-between items-center mb-4">
        <div className="text-sm text-muted-foreground">Tokens</div>
        <Badge variant="secondary">
          {CHAINS[chainId as keyof typeof CHAINS].name}
        </Badge>
      </div>
      <Card>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Symbol</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Address</TableHead>
                <TableHead className="text-right">Decimals</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {tokens.map((token: Token) => (
                <TableRow key={token.token_address}>
                  <TableCell className="font-semibold text-sm">
                    {token.symbol}
                  </TableCell>

                  <TableCell className="text-sm">{token.name}</TableCell>

                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {token.token_address}
                  </TableCell>

                  <TableCell className="text-right text-xs">
                    {token.decimals}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          <PaginationControls
            offset={offset}
            limit={limit}
            total={total}
            currentCount={tokens.length}
            onPrevious={() => setOffset((prev) => Math.max(0, prev - limit))}
            onNext={() => setOffset((prev) => prev + limit)}
          />
        </CardContent>
      </Card>
    </div>
  );
}
