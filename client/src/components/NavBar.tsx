import { Link, useLocation } from "react-router";
import { pages } from "@/utils/pages";

export default function NavBar() {
  const location = useLocation();
  return (
    <div className="m-8">
      {pages.map((page) => {
        return (
          <Link
            to={page.url}
            className={`${
              location.pathname === page.url
                ? "bg-green-500 text-blue-800 cursor-default"
                : "bg-blue-700 text-green-400 cursor-pointer"
            } pr-6 pl-6 pt-2 pb-2 ${!page.isFirstElement && "ml-2"}`}
          >
            {page.linkTitle}
          </Link>
        );
      })}
    </div>
  );
}
