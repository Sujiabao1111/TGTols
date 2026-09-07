export const dynamic = "force-static"

export default function TermsPage() {
  return (
    <main className="min-h-screen bg-white px-6 py-10 text-gray-900">
      <article className="mx-auto max-w-3xl space-y-5">
        <h1 className="text-3xl font-bold">Terms of Use</h1>
        <p>These terms govern your use of the PPNET gaming platform and Telegram Mini App.</p>
        <h2 className="text-xl font-semibold">Wallet payments</h2>
        <p>TON payments are signed by your wallet through TON Connect. An order is credited only after the server verifies the on-chain transaction, amount, destination and sender.</p>
        <h2 className="text-xl font-semibold">Acceptable use</h2>
        <p>You must provide accurate account information and may not attempt to reuse transactions, manipulate payment status or access another user&apos;s account.</p>
        <h2 className="text-xl font-semibold">Support</h2>
        <p>Contact support through the official Telegram channel linked in the application for questions about an order.</p>
        <p className="text-sm text-gray-500">Last updated: 2026-09-03</p>
      </article>
    </main>
  )
}
