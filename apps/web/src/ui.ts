export function el<K extends keyof HTMLElementTagNameMap>(tag: K, className = '', text = '') {
  const node = document.createElement(tag)
  node.className = className
  node.textContent = text
  return node
}

export function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
