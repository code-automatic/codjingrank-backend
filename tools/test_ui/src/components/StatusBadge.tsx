export default function StatusBadge({ status }: { status?: string }) {
  const cls = `status-badge status-${status?.replace(/ /g, '') || 'Unknown'}`
  return <span className={cls}>{status || 'Unknown'}</span>
}
