import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Field, FieldLabel } from '@/components/ui/field'
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group'
import { ShortenRequestWritable, zShortenRequestWritable } from '@/gen/api'
import { shortenUrlMutation } from '@/gen/api/@tanstack/react-query.gen'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { Loader2, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import GridiconsDomains from '~icons/gridicons/domains'
import LucideLink from '~icons/lucide/link'

type Props = {}

const AddURLDialog = (props: Props) => {
  const mutation = useMutation({
    ...shortenUrlMutation(),
    onSuccess: () => {
      toast.success('URL shortened successfully!')
    },
    onError: (error) => {
      toast.error(`Error shortening URL: ${error.errors && error.errors[0]?.message || error.detail}`)
    },
  })
  const [open, setOpen] = useState(false)
  const form = useForm({
    defaultValues: {
      original_url: '',
      domain: '',
    } as ShortenRequestWritable,
    validators: {
      onSubmit: zShortenRequestWritable,
      onBlur: zShortenRequestWritable,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({
        body: {
          original_url: value.original_url,
          domain: value.domain || undefined,
        }
      })
    },
  })
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4" />
          Add URL
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-106.25">
        <DialogHeader>
          <DialogTitle>Create Short Link</DialogTitle>
          <DialogDescription>
            Paste your long URL below to generate a short, shareable link.
          </DialogDescription>
        </DialogHeader>

        {/* 3. Form Render */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
          className="space-y-4"
        >
          {/* URL Field */}
          <form.Field name="original_url">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name} className={field.state.meta.errors.length ? "text-destructive" : ""}>
                  Destination URL
                </FieldLabel>
                <InputGroup>
                  <InputGroupAddon>
                    <LucideLink />
                  </InputGroupAddon>
                  <InputGroupInput
                    id={field.name}
                    placeholder="https://super-long-url.com/..."
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />

                  {/* Error Message */}
                  {field.state.meta.errors ? (
                    <p className="text-[0.8rem] font-medium text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  ) : null}
                </InputGroup>
              </Field>
            )}
          </form.Field>

          <form.Field name="domain">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name} className={field.state.meta.errors.length ? "text-destructive" : ""}>
                  Domain
                </FieldLabel>
                <InputGroup>
                  <InputGroupAddon>
                    <GridiconsDomains />
                  </InputGroupAddon>
                  <InputGroupInput
                    id={field.name}
                    disabled
                    placeholder="short.ly"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </InputGroup>
                {/* Error Message */}
                {field.state.meta.errors ? (
                  <p className="text-[0.8rem] font-medium text-destructive">
                    {field.state.meta.errors[0]?.message}
                  </p>
                ) : null}
              </Field>
            )}
          </form.Field>

          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => setOpen(false)}>
              Cancel
            </Button>

            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {isSubmitting ? "Shortening..." : "Create Link"}
                </Button>
              )}
            />
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default AddURLDialog