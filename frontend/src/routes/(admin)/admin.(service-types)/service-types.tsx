import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute(
  '/(admin)/admin/(service-types)/service-types',
)({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/(admin)/admin/(service-types)/service-types"!</div>
}
