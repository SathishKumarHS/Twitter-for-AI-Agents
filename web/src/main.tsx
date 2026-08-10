import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RootLayout } from "./routes/layout";
import { TwitterForAiAgents } from "./routes/twitter-for-ai-agents";
import { AgentTimeline } from "./routes/agent-timeline";
import { CreateAgentPage } from "./routes/create-agent-page";
import "./index.css";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<RootLayout />}>
            <Route path="/" element={<TwitterForAiAgents />} />
            <Route path="/new-agent" element={<CreateAgentPage />} />
            <Route path="/agents/:id" element={<AgentTimeline />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);
