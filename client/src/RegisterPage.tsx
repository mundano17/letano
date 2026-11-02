import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { zodResolver } from "@hookform/resolvers/zod"
import { cn } from "cn"
import { useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"
import apiFetch from "./lib/fetch"

const schema = z
  .object({
    email: z
      .string()
      .min(1, "Email is required")
      .regex(/^\S+@\S+\.\S+$/, "Invalid email"),
    initialPassword: z
      .string()
      .min(8, "Password must be a minimum of 8 characters")
      .regex(/[a-z]/, "Must contain at least one lowercase letter")
      .regex(/[A-Z]/, "Must contain at least one uppercase letter")
      .regex(/[0-9]/, "Must contain at least one number")
      .regex(/[^A-Za-z0-9]/, "Must contain at least one special character"),
    confirmPassword: z.string().min(1, "Confirm Password is required"),
  })
  .refine((data) => data.initialPassword === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  })

type schemaType = z.infer<typeof schema>

function RegisterForm({ className, ...props }: React.ComponentProps<"div">) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<schemaType>({
    resolver: zodResolver(schema),
  })

  interface registerData {
    Email: string
    Password: string
  }

  async function onSubmit(data: schemaType) {
    if (data.initialPassword != data.confirmPassword) {
      toast.error("Passwords must match")
      return
    }

    const regData: registerData = {
      Email: data.email,
      Password: data.confirmPassword,
    }

    const res = await apiFetch<typeof regData, unknown>(
      "/user/register",
      regData,
      "POST",
    )
    if (res.error) {
      toast.error(res.error)
      return
    }
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Register</CardTitle>
          <CardDescription>
            Enter your email below to register your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} noValidate>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  placeholder="m@example.com"
                  {...register("email")}
                />
                {errors.email && (
                  <FieldDescription className="text-destructive">
                    {errors.email.message}
                  </FieldDescription>
                )}
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="intialPassword">Password</FieldLabel>
                </div>
                <Input
                  id="intialPassword"
                  type="password"
                  {...register("initialPassword")}
                />
                {errors.initialPassword && (
                  <FieldDescription className="text-destructive">
                    {errors.initialPassword.message}
                  </FieldDescription>
                )}
                <div className="flex items-center">
                  <FieldLabel htmlFor="confirmPassword">
                    Confirm Password
                  </FieldLabel>
                </div>
                <Input
                  id="confirmPassword"
                  type="password"
                  {...register("confirmPassword")}
                />
                {errors.confirmPassword && (
                  <FieldDescription className="text-destructive">
                    {errors.confirmPassword.message}
                  </FieldDescription>
                )}
              </Field>
              <Button type="submit">Login</Button>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

export default function RegisterPage() {
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm">
        <RegisterForm />
      </div>
    </div>
  )
}
