"use client";

import { Button } from "@/components/ui/button";

interface PaginationControlsProps {
  offset: number;
  limit: number;
  total: number;
  currentCount: number;
  onPrevious: () => void;
  onNext: () => void;
}

export function PaginationControls({
  offset,
  limit,
  total,
  currentCount,
  onPrevious,
  onNext,
}: PaginationControlsProps) {
  const start = total === 0 ? 0 : offset + 1;
  const end = offset + currentCount;

  return (
    <div className="flex items-center justify-between mt-4">
      <div className="text-sm text-muted-foreground">
        Showing {start}-{end} of {total}
      </div>

      <div className="flex gap-2">
        <Button
          disabled={offset === 0}
          onClick={onPrevious}
          variant="secondary"
        >
          Previous
        </Button>

        <Button
          disabled={offset + limit >= total}
          onClick={onNext}
          variant="secondary"
        >
          Next
        </Button>
      </div>
    </div>
  );
}
