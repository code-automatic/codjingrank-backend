interface Props { message: string | null; onDismiss: () => void }
export default function ErrorBanner({ message, onDismiss }: Props) {
  if (!message) return null
  return (
    <div className="card" style={{ border: '1px solid #e53e3e', background: '#fff5f5' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="error">{message}</span>
        <button onClick={onDismiss} className="danger" style={{ padding: '2px 8px' }}>✕</button>
      </div>
    </div>
  )
}
