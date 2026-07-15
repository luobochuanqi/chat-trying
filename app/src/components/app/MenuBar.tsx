import { useTranslation } from "react-i18next";
import { useDispatch, useSelector } from "react-redux";
import {
  logout,
  selectAdmin,
  selectAuthenticated,
  selectUsername,
} from "@/store/auth.ts";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu.tsx";
import { Button } from "@/components/ui/button.tsx";
import router from "@/router.tsx";
import React from "react";
import { Image, MessageCircle, Shield, Store } from "lucide-react";
import Icon from "@/components/utils/Icon.tsx";

type MenuBarProps = {
  children: React.ReactNode;
  className?: string;
};

function MenuBar({ children, className }: MenuBarProps) {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const auth = useSelector(selectAuthenticated);
  const username = useSelector(selectUsername);
  const admin = useSelector(selectAdmin);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{children}</DropdownMenuTrigger>
      <DropdownMenuContent className={className} align={`end`}>
        {auth ? (
          <>
            <DropdownMenuLabel className={`username`}>
              {username}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => router.navigate("/")}>
              <Icon icon={<MessageCircle />} className={`w-4 h-4 mr-1.5`} />
              {t("bar.chat-full")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.navigate("/market")}>
              <Icon icon={<Store />} className={`w-4 h-4 mr-1.5`} />
              模型市场
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.navigate("/gallery")}>
              <Icon icon={<Image />} className={`w-4 h-4 mr-1.5`} />
              作品墙
            </DropdownMenuItem>
            {admin && (
              <DropdownMenuItem onClick={() => router.navigate("/admin")}>
                <Icon icon={<Shield />} className={`w-4 h-4 mr-1.5`} />
                {t("bar.admin-full")}
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Button
                size={`sm`}
                className={`action-button`}
                onClick={() => dispatch(logout())}
              >
                {t("logout")}
              </Button>
            </DropdownMenuItem>
          </>
        ) : (
          <DropdownMenuItem asChild>
            <Button
              size={`sm`}
              className={`h-max w-full cursor-pointer`}
              onClick={() => router.navigate("/login")}
            >
              {t("login")}
            </Button>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default MenuBar;
