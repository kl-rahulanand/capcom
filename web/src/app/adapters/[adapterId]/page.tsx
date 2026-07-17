import { AdapterDetail } from "@/components/adapter-detail"

export default async function AdapterPage({
  params,
}: {
  params: Promise<{ adapterId: string }>
}) {
  const { adapterId } = await params

  return <AdapterDetail adapterId={decodeURIComponent(adapterId)} />
}
