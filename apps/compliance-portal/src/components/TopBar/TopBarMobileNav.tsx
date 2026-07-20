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

import { Drawer } from "@base-ui/react/drawer";
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
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useLocation } from "react-router";

import { buildRequestAllContinueUrl } from "#/lib/auth/continueUrl";
import { useSignInDialog } from "#/lib/auth/signInDialogContext";
import { useSignOut } from "#/lib/auth/useSignOut";
import { useSubscribeDialog } from "#/lib/mailingList/subscribeDialogContext";

import type { TopBarMobileNav_identity$key } from "./__generated__/TopBarMobileNav_identity.graphql";
import { TOP_BAR_NAV_ITEMS } from "./navItems";
import { topBar, topBarMobileNav } from "./variants";

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
  const { openSignIn } = useSignInDialog();
  const { openSubscribe, isSubscribed, unsubscribe, isUnsubscribing } = useSubscribeDialog();
  const [signOut, isSigningOut] = useSignOut();
  const [open, setOpen] = useState(false);
  const identity = useFragment(topBarMobileNavFragment, identityKey);

  const barSlots = topBar();
  const slots = topBarMobileNav();

  const displayName = identity
    ? (identity.fullName.trim() || identity.email)
    : null;

  const close = () => setOpen(false);

  return (
    <Drawer.Root open={open} onOpenChange={setOpen} swipeDirection="right">
      <div className={barSlots.menuTrigger()}>
        <Drawer.Trigger
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
      <Drawer.Portal>
        <Drawer.Backdrop className={slots.backdrop()} />
        <Drawer.Viewport className={slots.viewport()}>
          <Drawer.Popup className={slots.popup()}>
            <Drawer.Content className={slots.content()}>
              <div className={slots.header()}>
                <Drawer.Title className={slots.title()}>
                  {t("topBar.menuTitle")}
                </Drawer.Title>
                <Drawer.Close
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
              </div>

              <nav className={slots.nav()}>
                {TOP_BAR_NAV_ITEMS.map(item => (
                  <Link
                    key={item.to}
                    to={item.to}
                    variant="ghost"
                    color="neutral"
                    size={3}
                    active={isActive(pathname, item.to)}
                    className="w-full justify-start"
                    onClick={close}
                  >
                    {t(item.labelKey)}
                  </Link>
                ))}
              </nav>

              <div className={slots.actions()}>
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
                          openSignIn({ continueTo: buildRequestAllContinueUrl() });
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
              </div>
            </Drawer.Content>
          </Drawer.Popup>
        </Drawer.Viewport>
      </Drawer.Portal>
    </Drawer.Root>
  );
}
