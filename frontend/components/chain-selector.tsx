"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type Chain = {
  chainId: number;
  name: string;
};

interface ChainSelectorProps {
  chains: Chain[];
  value: number;
  onChange: (chainId: number) => void;
  className?: string;
}

export function ChainSelector({
  chains,
  value,
  onChange,
  className,
}: ChainSelectorProps) {
  return (
    <Select
      value={value.toString()}
      onValueChange={(val) => onChange(Number(val))}
    >
      <SelectTrigger className={className ?? "w-[200px]"}>
        <SelectValue placeholder="Select chain" />
      </SelectTrigger>

      <SelectContent>
        {chains.map((chain) => (
          <SelectItem key={chain.chainId} value={chain.chainId.toString()}>
            {chain.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
