import { useState, useEffect } from "react";
import { graphql, useFragment } from "react-relay";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type PeopleSelector_organization$key } from "./__generated__/PeopleSelector_organization.graphql";

const peopleSelectorFragment = graphql`
  fragment PeopleSelector_organization on Organization {
    id
    peoples(first: 100, orderBy: { direction: ASC, field: FULL_NAME })
      @connection(key: "PeopleSelector_organization_peoples") {
      edges {
        node {
          id
          fullName
          primaryEmailAddress
        }
      }
    }
  }
`;

interface PeopleSelectorProps {
  organizationRef: PeopleSelector_organization$key;
  selectedPersonId: string | null;
  onSelect: (personId: string) => void;
  placeholder?: string;
  required?: boolean;
}

export default function PeopleSelector({
  organizationRef,
  selectedPersonId,
  onSelect,
  placeholder = "Select a person",
  required = false,
}: PeopleSelectorProps) {
  const organization = useFragment(peopleSelectorFragment, organizationRef);
  const [value, setValue] = useState<string>(selectedPersonId || "");

  useEffect(() => {
    if (selectedPersonId) {
      setValue(selectedPersonId);
    }
  }, [selectedPersonId]);

  const handleValueChange = (newValue: string) => {
    setValue(newValue);
    onSelect(newValue);
  };

  return (
    <Select value={value} onValueChange={handleValueChange} required={required}>
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {organization.peoples?.edges?.map(
          (edge) =>
            edge?.node && (
              <SelectItem key={edge.node.id} value={edge.node.id}>
                {edge.node.fullName} ({edge.node.primaryEmailAddress})
              </SelectItem>
            ),
        )}
      </SelectContent>
    </Select>
  );
}
