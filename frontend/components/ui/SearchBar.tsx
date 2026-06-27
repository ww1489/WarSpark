import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

export function SearchBar() {
  const [value, setValue] = useState("");
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const tag = value.trim().toUpperCase();
    if (!tag) return;
    const clean = tag.startsWith("#") ? tag : `#${tag}`;
    if (clean.startsWith("#") && clean.length >= 4) {
      navigate({ to: `/clans/${encodeURIComponent(clean)}` });
    }
    setValue("");
  };

  return (
    <form onSubmit={handleSubmit} className="relative">
      <input
        type="text"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder="输入部落标签 #2PP..."
        className="h-10 w-48 lg:w-64 pl-4 pr-4 rounded-tag border border-border bg-surface-page text-sm
          focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary
          placeholder:text-text-secondary/50"
      />
    </form>
  );
}
