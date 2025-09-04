import type { Page } from "@/types";
import Home from "@/pages/Home";
import About from "@/pages/About";

export const pages: Page[] = [
  {
    url: "/",
    linkTitle: "Home",
    isFirstElement: true,
    element: <Home />,
  },
  {
    url: "/about",
    linkTitle: "About us",
    element: <About />,
  },
];
