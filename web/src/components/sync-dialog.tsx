"use client"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

export type SyncDialogTarget = {
  id: string
  name: string
}

export type SyncDialogSubmit = {
  actor: string
  reason: string
}

type SyncDialogProps = {
  open: boolean
  targets: SyncDialogTarget[]
  pending?: boolean
  defaultReason: string
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: SyncDialogSubmit) => void
}

export function SyncDialog({
  open,
  targets,
  pending = false,
  defaultReason,
  onOpenChange,
  onSubmit,
}: SyncDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border border-[var(--hl)] bg-[var(--el)] shadow-[var(--shdw)] sm:max-w-[420px]">
        <form
          key={open ? defaultReason : "closed"}
          className="flex flex-col gap-4"
          onSubmit={(event) => {
            event.preventDefault()
            const formData = new FormData(event.currentTarget)
            const actor = String(formData.get("actor") ?? "").trim()
            const reason = String(formData.get("reason") ?? "").trim()
            if (!actor || !reason) {
              return
            }
            onSubmit({ actor, reason })
          }}
        >
          <DialogHeader>
            <DialogTitle>Re-import runtime state</DialogTitle>
            <DialogDescription>
              Record the operator and reason before calling the sync endpoint.
            </DialogDescription>
          </DialogHeader>

          <div className="rounded-lg border border-[var(--sl)] bg-[var(--sf)] px-3 py-2 font-hud text-[11px] text-[var(--mu)]">
            {targets.length === 1
              ? targets[0]?.name
              : `${targets.length} runtime instances`}
          </div>

          <label className="flex flex-col gap-2">
            <span className="capcom-eyebrow">Actor</span>
            <Input
              name="actor"
              defaultValue="local-operator"
              required
              autoComplete="username"
              className="font-hud text-[13px]"
            />
          </label>

          <label className="flex flex-col gap-2">
            <span className="capcom-eyebrow">Reason</span>
            <Textarea
              name="reason"
              defaultValue={defaultReason}
              required
              rows={3}
              className="resize-none font-hud text-[13px]"
            />
          </label>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={pending}
              className="shadow-[0_0_0_3px_var(--glow)] hover:brightness-[1.08]"
            >
              {pending ? "Re-importing" : "Re-import"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
