import Link from 'next/link'

export default function Home() {
  return (
    <main>
      <h1>Rusak</h1>
      <ul>
        <li><Link href="/good">Good</Link></li>
        <li><Link href="/bad">Bad</Link></li>
      </ul>
    </main>
  )
}
