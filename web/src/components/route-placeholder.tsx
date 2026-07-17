export function RoutePlaceholder({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <section className="flex flex-col gap-5">
      <div>
        <h1 className="text-[22px] font-bold leading-tight tracking-[-0.02em] text-[var(--tx)]">
          {title}
        </h1>
        <p className="mt-1 max-w-2xl text-[13px] text-[var(--mu)]">
          {description}
        </p>
      </div>

      <div className="capcom-placeholder">
        <div className="capcom-eyebrow">Stage 1 placeholder</div>
        <p className="mt-2 text-[13px] text-[var(--mu)]">
          Screen data and API-backed controls will be added in a later stage.
        </p>
      </div>
    </section>
  )
}
