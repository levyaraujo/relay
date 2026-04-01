export function formatReal(value: number | string): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(Number(value))
}

export function formatDateBR(date: string): string {
  return new Date(date).toLocaleDateString('pt-BR')
}
