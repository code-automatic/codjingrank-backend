export default function RepoNotes() {
  return (
    <div>
      <h2>GitHub / Webhook Notes</h2>
      <div className="info">
        <strong>Webhook Signature:</strong> GitHub signs webhook payloads with <code>X-Hub-Signature-256</code> using HMAC-SHA256.
        The <code>GITHUB_WEBHOOK_SECRET</code> must not be exposed to the browser.
        Webhooks should be tested from the terminal or GitHub's delivery page.
      </div>
      <div className="card">
        <h3>Webhook URL (local)</h3>
        <code>http://localhost:8084/webhooks/github</code>
      </div>
      <div className="card">
        <h3>Behind Tunnel</h3>
        <code>https://&lt;tunnel-domain&gt;/webhooks/github</code>
      </div>
      <div className="card">
        <h3>Test from Terminal</h3>
        <pre>{`bash scripts/send_github_webhook_test.sh`}</pre>
      </div>
      <div className="card">
        <h3>GitHub App Settings</h3>
        <ul style={{ paddingLeft: '1.5rem' }}>
          <li>Payload URL: your URL above</li>
          <li>Content type: application/json</li>
          <li>Secret: same as GITHUB_WEBHOOK_SECRET</li>
          <li>Events: Installation, Installation repositories, Repository</li>
        </ul>
      </div>
    </div>
  )
}
