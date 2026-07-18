import React, { useState, useEffect, useRef } from "react";
import { Wrench } from "lucide-react";
import { useTranslation } from "react-i18next";
import { ChatAction } from "@/components/home/assemblies/ChatAction.tsx";

interface Skill {
  name: string;
  description: string;
  parameters?: unknown;
}

interface Props {
  selected: string[];
  onChange: (selected: string[]) => void;
}

function SkillSelector({ selected, onChange }: Props) {
  const { t } = useTranslation();
  const [skills, setSkills] = useState<Skill[]>([]);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetch("/api/tools")
      .then((r) => r.json())
      .then((data) => {
        if (data.status) setSkills(Object.values(data.data));
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [open]);

  const toggle = (name: string) => {
    if (selected.includes(name)) {
      onChange(selected.filter((s) => s !== name));
    } else {
      onChange([...selected, name]);
    }
  };

  if (skills.length === 0) return null;

  return (
    <div className="relative" ref={ref}>
      <ChatAction
        active={selected.length > 0}
        text={t("tools") || "Tools"}
        onClick={() => setOpen(!open)}
      >
        <Wrench className={`h-4 w-4`} />
      </ChatAction>
      {open && (
        <div className="absolute bottom-full left-0 mb-2 w-64 rounded-md border bg-popover p-2 shadow-md z-50">
          <div className="text-xs font-medium text-muted-foreground mb-1.5 px-1">
            {t("select-tools") || "Select Tools"}
          </div>
          <div className="max-h-48 overflow-y-auto">
            {skills.map((skill) => (
              <label
                key={skill.name}
                className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-accent cursor-pointer text-sm"
              >
                <input
                  type="checkbox"
                  className="h-3.5 w-3.5 rounded border-input accent-primary"
                  checked={selected.includes(skill.name)}
                  onChange={() => toggle(skill.name)}
                />
                <span>{skill.description}</span>
              </label>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default SkillSelector;
