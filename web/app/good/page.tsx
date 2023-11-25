import { captureException } from '@sentry/nextjs'

async function getData() {
  const response = await fetch('http://localhost:8080/good')
  if (!response.ok) {
    captureException(new Error('Bad response'))
  }
  const data = await response.text()
  return data
}

export default async function Page() {
  const data = await getData()

  return (
    <>
      <h1>Good</h1>
      <p>{data}</p>
    </>
  )
}
