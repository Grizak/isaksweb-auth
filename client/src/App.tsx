import { Route, Routes, useLocation } from "react-router";
import Auth from "@/pages/Auth";
import Dashboard from "@/pages/Dashboard";
import Home from "@/pages/Home";
import { Helmet } from "react-helmet-async";
import { useEffect, useState } from "react";
import axios from "axios";
import type { BackendMetadataResponse } from "@/types";
import { delay } from "@/utils";

function App() {
  const location = useLocation();

  const [title, setTitle] = useState<string>("");
  const [metadata, setMetadata] = useState<{ [key: string]: string }>({});
  const [metaLoading, setMetaLoading] = useState<boolean>(true);

  useEffect(() => {
    const MAX_RETRIES = 5;
    const MAX_LONG_RETRIES = 15; // Total including one-minute retries
    let retryCount = 0;
    let isCancelled = false;

    const fetchResponse = async () => {
      if (isCancelled) return;

      try {
        const response = await axios.get<BackendMetadataResponse>(
          "/api/routeData",
          {
            params: {
              pathname: location.pathname,
            },
          }
        );

        if (isCancelled) return;

        setTitle(response.data.title ?? "");
        setMetadata(response.data.metadata ?? {});
        setMetaLoading(false);
      } catch (error) {
        if (isCancelled) return;

        console.error(`Fetch attempt ${retryCount + 1} failed:`, error);

        // Check if we should do fast exponential retries
        const shouldFastRetry =
          retryCount < MAX_RETRIES &&
          axios.isAxiosError(error) &&
          (!error.response || error.response.status >= 500);

        if (shouldFastRetry) {
          retryCount++;
          const waitTime = 800 * Math.pow(2, retryCount - 1);
          await delay(waitTime);
          await fetchResponse();
        } else {
          // Set defaults so app renders
          setTitle("Isaksweb Auth");
          setMetadata({});
          setMetaLoading(false);

          // Continue with slow retries if server might recover
          const shouldSlowRetry =
            retryCount < MAX_LONG_RETRIES &&
            axios.isAxiosError(error) &&
            (!error.response || error.response.status >= 500);

          if (shouldSlowRetry) {
            retryCount++;
            await delay(60 * 1000);
            await fetchResponse();
          }
        }
      }
    };

    fetchResponse();

    return () => {
      isCancelled = true; // Cancel ongoing requests on unmount/navigation
    };
  }, [location]);

  if (metaLoading) {
    return (
      <div className="bg-gradient-to-br from-blue-500 to-purple-600 min-h-screen flex items-center justify-center">
        <div className="text-center">
          {/* Spinner */}
          <div className="relative inline-block">
            <div className="w-20 h-20 border-4 border-white border-t-transparent rounded-full loading-animate-spin" />
          </div>

          {/* Loading text */}
          <h2 className="text-white text-2xl font-semibold mt-6 loading-animate-pulse">
            Loading...
          </h2>
        </div>
      </div>
    );
  }

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
