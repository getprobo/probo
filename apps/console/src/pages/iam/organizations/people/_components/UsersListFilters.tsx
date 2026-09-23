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

import { getMembershipRoles, peopleRoles } from "@probo/helpers";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

import {
  isUsersListKind,
  isUsersListRole,
  type UsersListKind,
  type UsersListProfileState,
  usersListProfileStates,
  useUsersListFilters,
} from "../_lib/useUsersListFilters";
import { usersList } from "../variants";

import { UsersListSearch } from "./UsersListSearch";

export function UsersListFilters() {
  const { t } = useTranslation();
  const { status, role, kind, setStatus, setRole, setKind } = useUsersListFilters();
  const { tools, filters, filter } = usersList();
  const allStatusesLabel = t("usersList.filters.allStatuses");
  const allRolesLabel = t("usersList.filters.allRoles");
  const allTypesLabel = t("usersList.filters.allTypes");

  return (
    <div className={tools()}>
      <UsersListSearch />
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={status}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setStatus(null);
                return;
              }
              if ((usersListProfileStates as readonly string[]).includes(value)) {
                setStatus(value as UsersListProfileState);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allStatusesLabel}
              aria-label={t("usersList.filters.status")}
            >
              {(value: UsersListProfileState | null) => (
                value != null
                  ? t(`usersList.filters.${value.toLowerCase()}`)
                  : allStatusesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allStatusesLabel}</SelectItem>
              {usersListProfileStates.map(state => (
                <SelectItem key={state} value={state}>
                  {t(`usersList.filters.${state.toLowerCase()}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={role}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setRole(null);
                return;
              }
              if (isUsersListRole(value)) {
                setRole(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allRolesLabel}
              aria-label={t("usersList.filters.role")}
            >
              {(value: string | null) => (
                value != null
                  ? getMembershipRoles(t).find(option => option.value === value)?.label ?? allRolesLabel
                  : allRolesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allRolesLabel}</SelectItem>
              {getMembershipRoles(t).map(({ value, label }) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={kind}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setKind(null);
                return;
              }
              if (isUsersListKind(value)) {
                setKind(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allTypesLabel}
              aria-label={t("usersList.filters.type")}
            >
              {(value: UsersListKind | null) => (
                value != null ? t(`personForm.kinds.${value}`) : allTypesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allTypesLabel}</SelectItem>
              {peopleRoles.map(peopleKind => (
                <SelectItem key={peopleKind} value={peopleKind}>
                  {t(`personForm.kinds.${peopleKind}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      </div>
    </div>
  );
}
