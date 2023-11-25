import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Rusak',
  description: 'Demonstrating sentry.io usage',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
