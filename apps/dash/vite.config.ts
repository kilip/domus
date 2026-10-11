import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import devtoolsJson from "vite-plugin-devtools-json";

export default defineConfig({
  plugins: [tailwindcss(), devtoolsJson(), reactRouter()],
  resolve: {
    tsconfigPaths: true,
  },
  preview: {
    host: "127.0.0.1",
  },
});
