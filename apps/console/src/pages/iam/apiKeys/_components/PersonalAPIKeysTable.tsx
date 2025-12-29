import { useTranslate } from "@probo/i18n";
import { formatDate } from "@probo/helpers";
import { Button, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";

export type PersonalAPIKeyRow = {
  id: string;
  name: string;
  createdAt: string;
  expiresAt: string;
  lastUsedAt: string | null;
};

export function PersonalAPIKeysTable(props: {
  keys: PersonalAPIKeyRow[];
  onRevoke: (key: { id: string; name: string }) => void;
  onShowToken: (key: { id: string; name: string }) => void;
  isShowingToken?: boolean;
}) {
  const { keys, onRevoke, onShowToken, isShowingToken } = props;
  const { __ } = useTranslate();
  const now = new Date();

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Name")}</Th>
          <Th>{__("Last used")}</Th>
          <Th>{__("Created")}</Th>
          <Th>{__("Expires")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {keys.map((k) => {
          const expired = new Date(k.expiresAt) < now;
          return (
            <Tr key={k.id}>
              <Td>
                <div className="font-medium text-txt-primary">{k.name}</div>
                <div className="text-xs text-txt-tertiary">
                  {expired ? __("Expired") : __("Active")}
                </div>
              </Td>
              <Td>
                <span className="text-sm text-txt-secondary">
                  {k.lastUsedAt ? formatDate(k.lastUsedAt) : "—"}
                </span>
              </Td>
              <Td>
                <span className="text-sm text-txt-secondary">
                  {formatDate(k.createdAt)}
                </span>
              </Td>
              <Td>
                <span className="text-sm text-txt-secondary">
                  {formatDate(k.expiresAt)}
                </span>
              </Td>
              <Td width={140} className="text-end">
                <div className="flex gap-2 justify-end">
                  <Button
                    variant="secondary"
                    onClick={() => onShowToken({ id: k.id, name: k.name })}
                    disabled={!!isShowingToken}
                  >
                    {__("Show")}
                  </Button>
                  <Button
                    variant="danger"
                    onClick={() => onRevoke({ id: k.id, name: k.name })}
                  >
                    {__("Revoke")}
                  </Button>
                </div>
              </Td>
            </Tr>
          );
        })}
      </Tbody>
    </Table>
  );
}
