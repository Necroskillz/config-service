import { Link } from '@tanstack/react-router';

export function AdminSidebar() {
  return (
    <aside className="w-64 p-4 flex flex-col gap-4">
      <Link to="/admin/variation-properties">Variation Properties</Link>
      <Link to="/admin/service-types">Service Types</Link>
    </aside>
  );
}
