import { useForm } from "@tanstack/react-form"
import { CalendarIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import z from "zod"
import { Button } from "../../components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "../../components/ui/card"
import { FieldGroup, FieldLabel } from "../../components/ui/field"
import { Input } from "../../components/ui/input"
import { Textarea } from "../../components/ui/textarea"

const urlFormSchema = z.object({
  destination: z.url('Must be a valid URL (e.g., https://example.com)')
    .min(1, 'Destination URL is required'),
  shortCode: z.string()
    .min(3, 'Short code must be at least 3 characters')
    .regex(/^[a-zA-Z0-9-_]+$/, 'Only letters, numbers, hyphens, and underscores allowed')
    .optional()
    .or(z.literal('')),
  title: z.string(),
  notes: z.string().max(500, 'Notes cannot exceed 500 characters').optional(),
  expiration: z.string().optional(),
})

type URLFormValues = z.infer<typeof urlFormSchema>

type Props = {}

const URLForm = (props: Props) => {
  const form = useForm({
    defaultValues: {
      destination: '',
      shortCode: '',
      title: '',
      notes: '',
      expiration: ''
    } as URLFormValues,
    validators: {
      onChange: urlFormSchema
    },
    onSubmit: (values) => {
      console.log('Form submitted with values:', values)

      toast.success('URL created successfully!')
    }
  })


  return (
    <Card className="w-full max-w-lg mx-auto">
      <CardHeader>
        <CardTitle>Create Short Link</CardTitle>
        <CardDescription>
          Enter your destination details below to generate a shortened URL.
        </CardDescription>
      </CardHeader>

      <form onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        form.handleSubmit()
      }}>
        <CardContent className="space-y-4">

          <FieldGroup>
            {/* Destination URL Field */}
            <form.Field
              name="destination"
              children={(field) => (
                <div className="space-y-2">
                  <FieldLabel htmlFor={field.name}>Destination URL <span className="text-red-500">*</span></FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="https://super-long-url.com/example"
                    className={field.state.meta.errors.length ? "border-red-500" : ""}
                  />
                  {field.state.meta.errors ? (
                    <p className="text-xs text-red-500 font-medium">
                      {field.state.meta.errors.join(', ')}
                    </p>
                  ) : null}
                </div>
              )}
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Custom Alias Field */}
              <form.Field
                name="shortCode"
                children={(field) => (
                  <div className="space-y-2">
                    <FieldLabel htmlFor={field.name}>Custom Alias (Optional)</FieldLabel>
                    <div className="relative">
                      <span className="absolute left-3 top-2.5 text-muted-foreground text-sm">/</span>
                      <Input
                        id={field.name}
                        name={field.name}
                        value={field.state.value || ''}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder="my-link"
                        className={`pl-6 ${field.state.meta.errors.length ? "border-red-500" : ""}`}
                      />
                    </div>
                    {field.state.meta.errors ? (
                      <p className="text-xs text-red-500 font-medium">
                        {field.state.meta.errors.join(', ')}
                      </p>
                    ) : null}
                  </div>
                )}
              />

              {/* Title Field */}
              <form.Field
                name="title"
                children={(field) => (
                  <div className="space-y-2">
                    <FieldLabel htmlFor={field.name}>Title (Optional)</FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value || ''}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="Marketing Campaign Q1"
                    />
                  </div>
                )}
              />
            </div>

            {/* Expiration Date Field */}
            <form.Field
              name="expiration"
              children={(field) => (
                <div className="space-y-2">
                  <FieldLabel htmlFor={field.name}>Expiration (Optional)</FieldLabel>
                  <div className="relative">
                    <Input
                      type="datetime-local"
                      id={field.name}
                      name={field.name}
                      value={field.state.value || ''}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className="block w-full"
                    />
                    {/* Decorative Icon */}
                    <CalendarIcon className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground pointer-events-none" />
                  </div>
                  <p className="text-[0.8rem] text-muted-foreground">
                    Link will return 404 after this date.
                  </p>
                </div>
              )}
            />

            {/* Notes Field */}
            <form.Field
              name="notes"
              children={(field) => (
                <div className="space-y-2">
                  <FieldLabel htmlFor={field.name}>Notes</FieldLabel>
                  <Textarea
                    id={field.name}
                    name={field.name}
                    value={field.state.value || ''}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="Internal usage notes..."
                    className="resize-none"
                  />
                  {field.state.meta.errors ? (
                    <p className="text-xs text-red-500 font-medium">
                      {field.state.meta.errors.join(', ')}
                    </p>
                  ) : null}
                </div>
              )}
            />
          </FieldGroup>
        </CardContent>

        <CardFooter className="flex justify-between border-t pt-6">
          <Button
            variant="outline"
            type="button"
            onClick={() => form.reset()}
          >
            Reset
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" disabled={!canSubmit}>
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Short URL'
                )}
              </Button>
            )}
          />
        </CardFooter>
      </form>
    </Card>
  )
}

export default URLForm