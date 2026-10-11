import { Breadcrumb, Input } from "antd";
import type { ReactNode } from "react";
import { Link } from "react-router";

export interface Crumb {
  title: ReactNode;
  /** Kosongkan untuk crumb terakhir (halaman saat ini). */
  to?: string;
}

interface ControlPanelProps {
  breadcrumbs: Crumb[];
  /** Kotak cari tampil hanya kalau placeholder diisi. */
  searchPlaceholder?: string;
  onSearch?: (value: string) => void;
  /** Tombol aksi di sisi kanan. Cukup satu `type="primary"` per layar. */
  actions?: ReactNode;
}

/** Control panel ala Odoo: breadcrumb, pencarian, dan tombol aksi. */
export function ControlPanel({
  breadcrumbs,
  searchPlaceholder,
  onSearch,
  actions,
}: ControlPanelProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3"
      style={{ borderBottom: "1px solid var(--domus-border-soft)" }}
    >
      <Breadcrumb
        className="min-w-0 sm:flex-1"
        items={breadcrumbs.map((crumb) => ({
          title: crumb.to ? (
            <Link to={crumb.to}>{crumb.title}</Link>
          ) : (
            crumb.title
          ),
        }))}
      />
      {searchPlaceholder && (
        <Input.Search
          allowClear
          className="w-full sm:w-72"
          placeholder={searchPlaceholder}
          onSearch={onSearch}
          onChange={(e) => onSearch?.(e.target.value)}
        />
      )}
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
}
