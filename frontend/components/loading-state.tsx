import { Card, CardHeader, CardContent } from "./ui/card";

export function LoadingState() {
  return (
    <div className="space-y-6 animate-pulse max-w-4xl mx-auto">
      {[1, 2].map((i) => (
        <Card key={i}>
          <CardHeader>
            <div className="h-4 w-40 bg-muted rounded" />
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[1, 2, 3, 4].map((j) => (
                <div key={j} className="h-4 w-full bg-muted rounded" />
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
