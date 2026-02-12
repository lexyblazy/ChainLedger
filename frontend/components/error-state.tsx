import { Card, CardContent, CardHeader } from "./ui/card";

export function ErrorState({ error, title }: { error: Error; title?: string }) {
  return (
    <Card>
      <CardHeader className="font-medium">
        {title || "Failed to load data."}
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        {error instanceof Error ? error.message : "Unknown error"}
      </CardContent>
    </Card>
  );
}
