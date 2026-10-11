import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ConfigProvider } from "antd";
import enUS from "antd/locale/en_US";
import idID from "antd/locale/id_ID";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from "react-router";
import { getCurrentLanguage, type Language } from "~/i18n";
import { getAntdTheme } from "~/shared/layout/theme";
import { useThemeMode } from "~/shared/layout/useThemeMode";

import type { Route } from "./+types/root";
import "./app.css";
import "~/i18n"; // inisialisasi i18next (side effect)

const ANTD_LOCALES = { id: idID, en: enUS } as const satisfies Record<
  Language,
  unknown
>;

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: false,
            refetchOnWindowFocus: false,
          },
        },
      }),
  );
  const mode = useThemeMode();
  const theme = useMemo(() => getAntdTheme(mode), [mode]);
  // Ikut berganti saat bahasa diubah (useTranslation subscribe ke i18n).
  const { i18n } = useTranslation();
  const language = getCurrentLanguage();

  // biome-ignore lint/correctness/useExhaustiveDependencies: `i18n.language` memicu sinkronisasi ke <html lang>
  useEffect(() => {
    document.documentElement.lang = language;
  }, [language, i18n.language]);

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider theme={theme} locale={ANTD_LOCALES[language]}>
        <Outlet />
      </ConfigProvider>
    </QueryClientProvider>
  );
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  const { t } = useTranslation();
  let message = t("error.title");
  let details = t("error.unexpected");
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message =
      error.status === 404 ? t("error.notFoundTitle") : t("error.generic");
    details =
      error.status === 404 ? t("error.notFound") : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="pt-16 p-4 container mx-auto">
      <h1>{message}</h1>
      <p>{details}</p>
      {stack && (
        <pre className="w-full p-4 overflow-x-auto">
          <code>{stack}</code>
        </pre>
      )}
    </main>
  );
}
