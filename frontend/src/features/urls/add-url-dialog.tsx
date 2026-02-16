import { Button } from "@/components/ui/button"
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog"
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field"
import {
	InputGroup,
	InputGroupAddon,
	InputGroupInput,
} from "@/components/ui/input-group"
import {
	type ShortenRequestWritable,
	zShortenRequestWritable,
} from "@/gen/api"
import { shortenUrlMutation } from "@/gen/api/@tanstack/react-query.gen"
import { useForm } from "@tanstack/react-form"
import { useMutation } from "@tanstack/react-query"
import { Loader2, Plus } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"
import FluentScreenCut20Regular from '~icons/fluent/screen-cut-20-regular'
import LucideLink from "~icons/lucide/link"
import LucideTextCursorInput from "~icons/lucide/text-cursor-input"

type Props = {}

const AddURLDialog = (props: Props) => {
	const mutation = useMutation({
		...shortenUrlMutation(),
		onSuccess: () => {
			toast.success("URL shortened successfully!")
			setOpen(false)
		},
		onError: (error) => {
			toast.error(
				`Error shortening URL: ${error.errors?.[0]?.message || error.detail}`,
			)
		},
	})
	const [open, setOpen] = useState(false)
	const form = useForm({
		defaultValues: {
			title: "",
			original_url: "",
			custom_alias: "",
		} as ShortenRequestWritable,
		validators: {
			onSubmit: zShortenRequestWritable,
			onBlur: zShortenRequestWritable,
		},
		onSubmit: async ({ value }) => {
			await mutation.mutateAsync({
				body: {
					title: value.title,
					original_url: value.original_url,
					custom_alias: value.custom_alias,
				},
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
								<FieldLabel
									htmlFor={field.name}
									className={
										field.state.meta.errors.length ? "text-destructive" : ""
									}
								>
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
								</InputGroup>
								{/* Error Message */}
								{field.state.meta.errors ? (
									<FieldDescription className="text-destructive">
										{field.state.meta.errors[0]?.message}
									</FieldDescription>
								) : null}
							</Field>
						)}
					</form.Field>

					{/* Title Field */}
					<form.Field name="title">
						{(field) => (
							<Field>
								<FieldLabel
									htmlFor={field.name}
									className={
										field.state.meta.errors.length ? "text-destructive" : ""
									}
								>
									Title
								</FieldLabel>
								<InputGroup>
									<InputGroupAddon>
										<LucideTextCursorInput />
									</InputGroupAddon>
									<InputGroupInput
										id={field.name}
										placeholder="My Awesome Link"
										value={field.state.value}
										onBlur={field.handleBlur}
										onChange={(e) => field.handleChange(e.target.value)}
									/>
								</InputGroup>
								{/* Error Message */}
								{field.state.meta.errors ? (
									<FieldDescription className="text-destructive">
										{field.state.meta.errors[0]?.message}
									</FieldDescription>
								) : null}
							</Field>
						)}
					</form.Field>
					{/* Custom Alias Field */}
					<form.Field name="custom_alias">
						{(field) => (
							<Field>
								<FieldLabel
									htmlFor={field.name}
									className={
										field.state.meta.errors.length ? "text-destructive" : ""
									}
								>
									Custom Alias (optional)
								</FieldLabel>
								<InputGroup>
									<InputGroupAddon>
										<FluentScreenCut20Regular />
									</InputGroupAddon>
									<InputGroupInput
										id={field.name}
										placeholder="my-link"
										value={field.state.value || ""}
										onBlur={field.handleBlur}
										onChange={(e) => field.handleChange(e.target.value)}
									/>
								</InputGroup>
								<FieldDescription>
									Choose a custom short code (max 20 characters)
								</FieldDescription>
								{/* Error Message */}
								{field.state.meta.errors ? (
									<FieldDescription className="text-destructive">
										{field.state.meta.errors[0]?.message}
									</FieldDescription>
								) : null}
							</Field>
						)}
					</form.Field>


					<DialogFooter>
						<Button
							variant="outline"
							type="button"
							onClick={() => setOpen(false)}
						>
							Cancel
						</Button>

						<form.Subscribe
							selector={(state) => [state.canSubmit, state.isSubmitting]}
						>
							{([canSubmit, isSubmitting]) => (
								<Button type="submit" disabled={!canSubmit || isSubmitting}>
									{isSubmitting && (
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
									)}
									{isSubmitting ? "Shortening..." : "Create Link"}
								</Button>
							)}
						</form.Subscribe>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	)
}

export default AddURLDialog
