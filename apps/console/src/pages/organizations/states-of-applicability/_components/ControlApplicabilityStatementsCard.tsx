import { graphql } from "relay-runtime";
import {
    Card,
    IconPlusLarge,
    Button,
    Tr,
    Td,
    Table,
    Thead,
    Tbody,
    Th,
    IconChevronDown,
    IconTrashCan,
    TrButton,
    Badge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { ControlApplicabilityStatementsCardFragment$key } from "/__generated__/core/ControlApplicabilityStatementsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState, useEffect } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { SelectStateOfApplicabilityDialog } from "./SelectStateOfApplicabilityDialog";
import clsx from "clsx";

const controlApplicabilityStatementsCardFragment = graphql`
    fragment ControlApplicabilityStatementsCardFragment on StateOfApplicabilityControl {
        id
        stateOfApplicabilityId
        controlId
        stateOfApplicability {
            id
            name
        }
        applicability
        justification
    }
`;

type AttachMutation<Params> = (p: {
    variables: {
        input: {
            stateOfApplicabilityId: string;
            applicability: boolean;
            justification: string | null;
        } & Params;
        connections: string[];
    };
}) => void;

type DetachMutation = (p: {
    variables: {
        input: {
            applicabilityStatementId: string;
        };
        connections: string[];
    };
}) => void;

type Props<Params> = {
    statesOfApplicability: readonly (ControlApplicabilityStatementsCardFragment$key & {
        id: string;
    })[];
    params: Params;
    disabled?: boolean;
    connectionId: string;
    onAttach: AttachMutation<Params>;
    onDetach: DetachMutation;
    variant?: "card" | "table";
    readOnly?: boolean;
};

export function ControlApplicabilityStatementsCard<Params>(props: Props<Params>) {
    const { __ } = useTranslate();

    const [limit, setLimit] = useState<number | null>(
        props.variant === "card" ? 4 : null,
    );

    const [linkedInfo, setLinkedInfo] = useState<
        { applicabilityStatementId: string; stateOfApplicabilityId: string }[]
    >([]);

    const statesOfApplicability = useMemo(() => {
        return limit
            ? props.statesOfApplicability.slice(0, limit)
            : props.statesOfApplicability;
    }, [props.statesOfApplicability, limit]);

    const showMoreButton =
        limit !== null && props.statesOfApplicability.length > limit;
    const variant = props.variant ?? "table";

    const linkedData = linkedInfo;

    const onAttach = (
        stateOfApplicabilityId: string,
        applicability: boolean,
        justification: string | null,
    ) => {
        props.onAttach({
            variables: {
                input: {
                    stateOfApplicabilityId,
                    applicability,
                    justification,
                    ...props.params,
                },
                connections: [props.connectionId],
            },
        });
    };

    const onDetach = (applicabilityStatementId: string) => {
        props.onDetach({
            variables: {
                input: {
                    applicabilityStatementId,
                },
                connections: [props.connectionId],
            },
        });
    };

    const Wrapper = variant === "card" ? Card : "div";

    return (
        <Wrapper padded className="space-y-[10px]">
            {props.statesOfApplicability.map((soa, idx) => (
                <LinkedInfoExtractor
                    key={idx}
                    fKey={soa}
                    onExtracted={(info) => {
                        setLinkedInfo((prev) => {
                            const exists = prev.some(
                                (p) =>
                                    p.applicabilityStatementId ===
                                    info.applicabilityStatementId,
                            );
                            return exists ? prev : [...prev, info];
                        });
                    }}
                />
            ))}
            {variant === "card" && (
                <div className="flex justify-between">
                    <div className="text-lg font-semibold">
                        {__("States of Applicability")}
                    </div>
                    {!props.readOnly && (
                        <SelectStateOfApplicabilityDialog
                            connectionId={props.connectionId}
                            disabled={props.disabled}
                            linkedStatesOfApplicability={linkedData}
                            onLink={onAttach}
                            onUnlink={onDetach}
                        >
                            <Button variant="tertiary" icon={IconPlusLarge}>
                                {__("Link state of applicability")}
                            </Button>
                        </SelectStateOfApplicabilityDialog>
                    )}
                </div>
            )}
            <Table className={clsx(variant === "card" && "bg-invert")}>
                <Thead>
                    <Tr>
                        <Th>{__("Name")}</Th>
                        <Th>{__("Applicability")}</Th>
                        <Th>{__("Justification")}</Th>
                        {!props.readOnly && <Th></Th>}
                    </Tr>
                </Thead>
                <Tbody>
                    {statesOfApplicability.length === 0 && (
                        <Tr>
                            <Td
                                colSpan={props.readOnly ? 3 : 4}
                                className="text-center text-txt-secondary"
                            >
                                {__("No states of applicability linked")}
                            </Td>
                        </Tr>
                    )}
                    {statesOfApplicability.map((soa) => (
                        <ApplicabilityStatementRow
                            key={soa.id}
                            fKey={soa}
                            onClick={onDetach}
                            readOnly={props.readOnly}
                        />
                    ))}
                    {variant === "table" && !props.readOnly && (
                        <SelectStateOfApplicabilityDialog
                            connectionId={props.connectionId}
                            disabled={props.disabled}
                            linkedStatesOfApplicability={linkedData}
                            onLink={onAttach}
                            onUnlink={onDetach}
                        >
                            <TrButton colspan={4} icon={IconPlusLarge}>
                                {__("Link state of applicability")}
                            </TrButton>
                        </SelectStateOfApplicabilityDialog>
                    )}
                </Tbody>
            </Table>
            {showMoreButton && (
                <Button
                    variant="tertiary"
                    icon={IconChevronDown}
                    onClick={() => setLimit(null)}
                >
                    {sprintf(
                        __("Show %d more"),
                        props.statesOfApplicability.length - limit!,
                    )}
                </Button>
            )}
        </Wrapper>
    );
}

function LinkedInfoExtractor({
    fKey,
    onExtracted,
}: {
    fKey: ControlApplicabilityStatementsCardFragment$key;
    onExtracted: (info: {
        applicabilityStatementId: string;
        stateOfApplicabilityId: string;
    }) => void;
}) {
    const data = useFragment(controlApplicabilityStatementsCardFragment, fKey);

    useEffect(() => {
        onExtracted({
            applicabilityStatementId: data.id,
            stateOfApplicabilityId: data.stateOfApplicabilityId,
        });
    }, [data.id, data.stateOfApplicabilityId, onExtracted]);

    return null;
}

function ApplicabilityStatementRow({
    fKey,
    onClick,
    readOnly,
}: {
    fKey: ControlApplicabilityStatementsCardFragment$key & {
        id: string;
    };
    onClick: (applicabilityStatementId: string) => void;
    readOnly?: boolean;
}) {
    const statement = useFragment(
        controlApplicabilityStatementsCardFragment,
        fKey,
    );
    const organizationId = useOrganizationId();
    const { __ } = useTranslate();

    return (
        <Tr
            to={`/organizations/${organizationId}/states-of-applicability/${statement.stateOfApplicabilityId}`}
        >
            <Td>{statement.stateOfApplicability.name}</Td>
            <Td>
                <Badge variant={statement.applicability ? "success" : "danger"}>
                    {statement.applicability
                        ? __("Applicable")
                        : __("Not Applicable")}
                </Badge>
            </Td>
            <Td>{statement.justification || "-"}</Td>
            {!readOnly && (
                <Td noLink width={50} className="text-end">
                    <Button
                        variant="secondary"
                        onClick={() => onClick(statement.id)}
                        icon={IconTrashCan}
                    >
                        {__("Unlink")}
                    </Button>
                </Td>
            )}
        </Tr>
    );
}

