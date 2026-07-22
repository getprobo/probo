// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import {
  BellIcon,
  BellRingingIcon,
  ListIcon,
  LockSimpleIcon,
  SignOutIcon,
  XIcon,
} from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { Drawer } from "@probo/ui/src/v2/Drawer/Drawer";
import { DrawerBody } from "@probo/ui/src/v2/Drawer/DrawerBody";
import { DrawerClose } from "@probo/ui/src/v2/Drawer/DrawerClose";
import { DrawerFooter } from "@probo/ui/src/v2/Drawer/DrawerFooter";
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerPopup } from "@probo/ui/src/v2/Drawer/DrawerPopup";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { DrawerTrigger } from "@probo/ui/src/v2/Drawer/DrawerTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useLocation } from "react-router";

import { buildRequestAllContinueUrl, redirectToInitiate } from "#/lib/auth/continueUrl";
import { useSignOut } from "#/lib/auth/useSignOut";
import { useLocalizedPath } from "#/lib/i18n/useLocale";
import { useSubscribeDialog } from "#/lib/mailingList/subscribeDialogContext";

import type { TopBarMobileNav_identity$key } from "./__generated__/TopBarMobileNav_identity.graphql";
import { LocaleSelect } from "./LocaleSelect";
import { TOP_BAR_NAV_ITEMS } from "./navItems";
import { topBar } from "./variants";

const topBarMobileNavFragment = graphql`
  fragment TopBarMobileNav_identity on Identity {
    fullName
    email
  }
`;

interface TopBarMobileNavProps {
  identityKey: TopBarMobileNav_identity$key | null;
}

function isActive(pathname: string, to: string): boolean {
  return pathname === to || pathname.startsWith(`${to}/`);
}

// Burger trigger + right-edge drawer listing the same nav and account actions as
// the desktop TopBar.
export function TopBarMobileNav({ identityKey }: TopBarMobileNavProps) {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const { openSubscribe, isSubscribed, unsubscribe, isUnsubscribing } = useSubscribeDialog();
  const [signOut, isSigningOut] = useSignOut();
  const [open, setOpen] = useState(false);
  // Select menus portal to body at z-3 by default; mount them on the drawer
  // popup so they stack inside the modal layer (z-5) instead of under it.
  const drawerPopupRef = useRef<HTMLDivElement>(null);
  const identity = useFragment(topBarMobileNavFragment, identityKey);
  const localizedPath = useLocalizedPath();

  const barSlots = topBar();

  const displayName = identity
    ? (identity.fullName.trim() || identity.email)
    : null;

  const close = () => setOpen(false);

  return (
    <Drawer open={open} onOpenChange={setOpen} swipeDirection="right">
      <div className={barSlots.menuTrigger()}>
        <DrawerTrigger
          render={(
            <IconButton
              variant="ghost"
              color="neutral"
              size={2}
              aria-label={t("topBar.openMenu")}
            >
              <ListIcon />
            </IconButton>
          )}
        />
      </div>
      <DrawerPopup side="right" ref={drawerPopupRef}>
        <DrawerHeader>
          <DrawerTitle>{t("topBar.menuTitle")}</DrawerTitle>
          <DrawerClose
            render={(
              <IconButton
                variant="ghost"
                color="neutral"
                size={2}
                aria-label={t("topBar.closeMenu")}
              >
                <XIcon />
              </IconButton>
            )}
          />
        </DrawerHeader>

        <DrawerBody>
          <nav className="contents">
            {TOP_BAR_NAV_ITEMS.map((item) => {
              const to = localizedPath(item.to);
              return (
                <Link
                  key={item.to}
                  to={to}
                  variant="ghost"
                  color="neutral"
                  size={3}
                  active={isActive(pathname, to)}
                  className="w-full justify-start"
                  onClick={close}
                >
                  {t(item.labelKey)}
                </Link>
              );
            })}
          </nav>
        </DrawerBody>

        <DrawerFooter>
          <div className="w-full pb-2">
            <LocaleSelect
              persist={identity != null}
              onLocaleChange={close}
              portalContainer={drawerPopupRef}
            />
          </div>
          {identity == null
            ? (
                <Button
                  variant="solid"
                  color="neutral"
                  highContrast
                  size={3}
                  className="w-full"
                  iconStart={<LockSimpleIcon />}
                  onClick={() => {
                    close();
                    redirectToInitiate(buildRequestAllContinueUrl());
                  }}
                >
                  {t("topBar.getAccess")}
                </Button>
              )
            : (
                <>
                  {displayName
                    ? (
                        <div className="flex flex-col gap-0.5 px-1 pb-2">
                          <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
                            {displayName}
                          </Text>
                          <Text size={1} color="faint" className="truncate">
                            {identity.email}
                          </Text>
                        </div>
                      )
                    : null}
                  <Button
                    variant="ghost"
                    color={isSubscribed ? "green" : "gold"}
                    size={3}
                    className="w-full justify-start"
                    iconStart={isSubscribed ? <BellRingingIcon /> : <BellIcon />}
                    disabled={isUnsubscribing}
                    onClick={() => {
                      close();
                      if (isSubscribed) {
                        void unsubscribe();
                        return;
                      }
                      openSubscribe();
                    }}
                  >
                    {isSubscribed ? t("userMenu.subscribed") : t("userMenu.subscribe")}
                  </Button>
                  <Button
                    variant="ghost"
                    color="red"
                    size={3}
                    className="w-full justify-start"
                    iconStart={<SignOutIcon />}
                    disabled={isSigningOut}
                    onClick={() => {
                      close();
                      void signOut();
                    }}
                  >
                    {t("userMenu.signOut")}
                  </Button>
                </>
              )}
        </DrawerFooter>
      </DrawerPopup>
    </Drawer>
  );
}
