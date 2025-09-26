import { Route, Routes, useLocation } from "react-router";
import Auth from "@/pages/Auth";
import Dashboard from "@/pages/Dashboard";
import Home from "@/pages/Home";
import { Helmet } from "react-helmet-async";
import { useEffect, useState } from "react";
import axios from "axios";
import type { BackendMetadataResponse } from "@/types";

function App() {
  const location = useLocation();

  const [title, setTitle] = useState<string>("");
  const [metadata, setMetadata] = useState<{ [key: string]: string }>({});

  useEffect(() => {
    (async () => {
      const response = await axios.get<BackendMetadataResponse>(
        "/api/routeData",
        {
          params: {
            location,
          },
        }
      );

      setTitle(response.data.title);
      setMetadata(response.data.metadata);
    })();
  }, [location]);

  return (
    <>
      <Helmet>
        <title>{title}</title>
        {Object.entries(metadata).map(([key, value]) => {
          return <meta content={value} name={key} />;
        })}
      </Helmet>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Auth state="login" />} />
        <Route path="/register" element={<Auth state="register" />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </>
  );
}

export default App;
