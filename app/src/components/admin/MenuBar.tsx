import { useDispatch, useSelector } from "react-redux";
import { closeMenu, selectMenu } from "@/store/menu.ts";
import React, { useMemo } from "react";
import {
  CopyCheck,
  FileClock,
  Gauge,
  Settings,
  Users,
} from "lucide-react";
import router from "@/router.tsx";
import { useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { mobile } from "@/utils/device.ts";
import { cn } from "@/components/ui/lib/utils.ts";

type MenuItemProps = {
  title: string;
  icon: React.ReactNode;
  path: string;
};

function MenuItem({ title, icon, path }: MenuItemProps) {
  const location = useLocation();
  const dispatch = useDispatch();
  const active = useMemo(
    () =>
      location.pathname === `/admin${path}` ||
      location.pathname + "/" === `/admin${path}`,
    [location.pathname, path],
  );

  const redirect = async () => {
    if (mobile) dispatch(closeMenu());
    await router.navigate(`/admin${path}`);
  };

  return (
    <div className={cn("menu-item", active && "active")} onClick={redirect}>
      <div className={`menu-item-icon`}>{icon}</div>
      <div className={`menu-item-title`}>{title}</div>
    </div>
  );
}

function MenuBar() {
  const { t } = useTranslation();
  const open = useSelector(selectMenu);
  return (
    <div className={cn("admin-menu", open && "open")}>
      <MenuItem title={t("admin.dashboard")} icon={<Gauge />} path={"/"} />
      <MenuItem title={t("admin.user")} icon={<Users />} path={"/users"} />
      <MenuItem title={"作品审核"} icon={<CopyCheck />} path={"/gallery"} />
      <MenuItem
        title={t("admin.settings")}
        icon={<Settings />}
        path={"/system"}
      />
      <MenuItem
        title={t("admin.logger.title")}
        icon={<FileClock />}
        path={"/logger"}
      />
    </div>
  );
}

export default MenuBar;
