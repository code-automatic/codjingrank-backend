interface Props { value: string; onChange: (v: string) => void; readOnly?: boolean; label?: string }
export default function CodeEditor({ value, onChange, readOnly, label }: Props) {
  return (
    <div className="form-group">
      {label && <label>{label}</label>}
      <textarea value={value} onChange={e => onChange(e.target.value)} readOnly={readOnly} style={{ fontFamily: 'monospace', minHeight: 150 }} />
    </div>
  )
}
