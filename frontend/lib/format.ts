export function formatTokenBalance(raw: string, decimals: number): string {
  try {
    const value = BigInt(raw);
    const isNegative = value < BigInt(0);

    const absValue = isNegative ? -value : value;
    const divisor = BigInt(10) ** BigInt(decimals);

    const whole = absValue / divisor;
    const fraction = absValue % divisor;

    const wholeStr = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");

    const fractionStr = fraction.toString().padStart(decimals, "0").slice(0, 4);

    return `${isNegative ? "-" : ""}${wholeStr}.${fractionStr}`;
  } catch {
    return "0";
  }
}

export const formatDate = (date: string) => {
  return new Date(date).toLocaleString(undefined, {
    year: "2-digit",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};
