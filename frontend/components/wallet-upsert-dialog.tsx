"use client";

import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Pencil } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

import { makeRequest } from "@/lib/api";

interface Wallet {
  address: string;
  label?: string;
  entity_type?: string;
}

interface WalletUpsertDialogProps {
  chainId: number;
  wallet?: Wallet;
}

export function WalletUpsertDialog({
  chainId,
  wallet,
}: WalletUpsertDialogProps) {
  const queryClient = useQueryClient();

  const isEdit = Boolean(wallet);

  const [open, setOpen] = useState(false);
  const [address, setAddress] = useState(wallet?.address ?? "");
  const [label, setLabel] = useState(wallet?.label ?? "");
  const [entityType, setEntityType] = useState(wallet?.entity_type ?? "");

  // Reset state when editing a different wallet
  useEffect(() => {
    if (wallet) {
      setAddress(wallet.address);
      setLabel(wallet.label ?? "");
      setEntityType(wallet.entity_type ?? "");
    } else {
      setAddress("");
      setLabel("");
      setEntityType("");
    }
  }, [wallet]);

  const mutation = useMutation({
    mutationFn: () =>
      makeRequest("/wallets", {
        method: "POST",
        body: JSON.stringify({
          address,
          chain_id: chainId,
          label,
          entity_type: entityType,
        }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["wallets", chainId],
      });

      setOpen(false);
    },
  });

  useEffect(() => {
    if (open) {
      mutation.reset();
    }
  }, [open, mutation]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant={isEdit ? "ghost" : "default"} size="sm">
          {isEdit ? <Pencil className="h-4 w-4" /> : "Add Wallet"}
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Wallet" : "Add Wallet"}</DialogTitle>
        </DialogHeader>

        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            mutation.mutate();
          }}
        >
          <div className="space-y-2">
            <Label>Address</Label>
            <Input
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              required
              disabled={isEdit}
            />
          </div>

          <div className="space-y-2">
            <Label>Label</Label>
            <Input value={label} onChange={(e) => setLabel(e.target.value)} />
          </div>

          <div className="space-y-2">
            <Label>Entity Type</Label>
            <Input
              value={entityType}
              onChange={(e) => setEntityType(e.target.value)}
            />
          </div>

          {mutation.isError && (
            <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {(mutation.error as Error).message}
            </div>
          )}

          <Button
            type="submit"
            className="w-full"
            disabled={mutation.isPending}
          >
            {mutation.isPending ? "Saving..." : isEdit ? "Update" : "Create"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
