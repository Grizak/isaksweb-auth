import { Route, Routes } from "react-router";
import NavBar from "@/components/NavBar";
import { pages } from "@/utils/pages";

function App() {
  return (
    <>
      <NavBar />
      <Routes>
        {pages.map((page) => (
          <Route element={page.element} path={page.url} />
        ))}
      </Routes>
    </>
  );
}

export default App;
