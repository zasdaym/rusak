import { captureException } from "@sentry/react"

async function getData() {
  const response = await fetch('http://localhost:8080/bad')
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
      <h1>Bad</h1>
      <p>{data}</p>
    </>
  )
}
