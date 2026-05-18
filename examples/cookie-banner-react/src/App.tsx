import { useCallback, useState } from "react";
import { ConfigTab } from "./tabs/ConfigTab";
import { ThemedBannerTab } from "./tabs/ThemedBannerTab";
import { HeadlessTab } from "./tabs/HeadlessTab";
import { DebugTab } from "./tabs/DebugTab";

export interface EventEntry {
  time: string;
  type: string;
  detail: unknown;
}

const tabs = [
  { id: "config", label: "Config" },
  { id: "themed", label: "Themed Banner" },
  { id: "headless", label: "Headless" },
  { id: "debug", label: "Debug" },
] as const;

type TabId = (typeof tabs)[number]["id"];

export function App() {
  const [activeTab, setActiveTab] = useState<TabId>("config");
  const [themedEvents, setThemedEvents] = useState<EventEntry[]>([]);
  const [headlessEvents, setHeadlessEvents] = useState<EventEntry[]>([]);

  const pushThemedEvent = useCallback((type: string, detail: unknown) => {
    setThemedEvents((prev) => [
      { time: new Date().toISOString(), type, detail },
      ...prev,
    ]);
  }, []);

  const pushHeadlessEvent = useCallback((type: string, detail: unknown) => {
    setHeadlessEvents((prev) => [
      { time: new Date().toISOString(), type, detail },
      ...prev,
    ]);
  }, []);

  return (
    <div style={{ fontFamily: "system-ui, sans-serif", maxWidth: 900, margin: "0 auto", padding: 24 }}>
      <h1 style={{ marginBottom: 4 }}>
        @probo/cookie-banner
      </h1>
      <p style={{ color: "#666", marginTop: 0, marginBottom: 24 }}>
        SDK example &mdash; React
      </p>

      <nav
        style={{
          display: "flex",
          gap: 0,
          borderBottom: "2px solid #ddd",
          marginBottom: 24,
        }}
      >
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            style={{
              padding: "8px 16px",
              border: "none",
              borderBottom:
                activeTab === tab.id ? "2px solid #333" : "2px solid transparent",
              background: "none",
              fontWeight: activeTab === tab.id ? "bold" : "normal",
              cursor: "pointer",
              marginBottom: -2,
              fontSize: 14,
            }}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {activeTab === "config" && <ConfigTab />}
      {activeTab === "themed" && <ThemedBannerTab events={themedEvents} pushEvent={pushThemedEvent} />}
      {activeTab === "headless" && <HeadlessTab events={headlessEvents} pushEvent={pushHeadlessEvent} />}
      {activeTab === "debug" && <DebugTab />}
    </div>
  );
}
