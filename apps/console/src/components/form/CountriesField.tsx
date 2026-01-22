import { countries, type CountryCode, getCountryName, getCountryOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, IconCrossLargeX, Input } from "@probo/ui";
import { clsx } from "clsx";
import { useEffect, useState } from "react";
import { type Control, Controller, type FieldPath, type FieldValues } from "react-hook-form";

type Props<T extends FieldValues = FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  disabled?: boolean;
};

export function CountriesField<T extends FieldValues = FieldValues>({ control, name, disabled }: Props<T>) {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <CountriesFieldInput
          value={field.value ?? []}
          onValueChange={field.onChange}
          disabled={disabled}
        />
      )}
    />
  );
}

type CountriesFieldInputProps = {
  value: string[];
  onValueChange: (value: string[]) => void;
  disabled?: boolean;
};

function CountriesFieldInput(props: CountriesFieldInputProps) {
  const { __ } = useTranslate();
  const [animateBadge, setAnimateBadge] = useState(false);

  const addCountry = (code: string) => {
    setAnimateBadge(true);
    props.onValueChange([...props.value, code]);
  };

  const removeCountry = (code: string) => {
    setAnimateBadge(true);
    props.onValueChange(props.value.filter(v => v !== code));
  };

  return (
    <div className={clsx(props.value.length > 0 ? "space-y-4" : "")}>
      {props.value.length > 0 && (
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            {props.value.map(countryCode => (
              <Badge asChild size="md" key={countryCode}>
                <button
                  onClick={() => removeCountry(countryCode)}
                  type="button"
                  disabled={props.disabled}
                  className={clsx(
                    "hover:bg-subtle-hover cursor-pointer",
                    props.disabled && "opacity-50 cursor-not-allowed",
                    animateBadge
                    && "starting:opacity-0 starting:w-0 w-max transition-all duration-500 starting:bg-accent",
                  )}
                >
                  {getCountryName(__, countryCode as CountryCode)}
                  <div className="w-0 overflow-hidden group-hover:w-4 duration-200">
                    <IconCrossLargeX size={12} />
                  </div>
                </button>
              </Badge>
            ))}
          </div>
        </div>
      )}
      {!props.disabled && (
        <CountryInput
          availableCountries={countries.filter(
            (c: CountryCode) => !props.value.includes(c),
          )}
          onAdd={addCountry}
        />
      )}
    </div>
  );
}

type CountryInputProps = {
  availableCountries: readonly CountryCode[];
  onAdd: (code: string) => void;
};

function CountryInput({ availableCountries, onAdd }: CountryInputProps) {
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const countryOptions = getCountryOptions(__);

  useEffect(() => {
    document.body.style.overflow = isOpen ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);

  const filteredCountries = countryOptions
    .filter((option: { value: string; label: string }) => availableCountries.includes(option.value as CountryCode))
    .filter((option: { value: string; label: string }) =>
      option.label.toLowerCase().includes(search.toLowerCase()),
    );

  const handleCountryClick = (value: string) => {
    onAdd(value);
    setSearch("");
  };

  return (
    <div className="relative">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Input
            type="text"
            value={search}
            onChange={e => setSearch(e.target.value)}
            onFocus={() => setIsOpen(true)}
            placeholder={__("Search and add countries...")}
            className="w-full pr-8"
          />
          <button
            type="button"
            onClick={() => setIsOpen(!isOpen)}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-txt-secondary hover:text-txt-primary transition-colors"
          >
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>
        </div>
      </div>

      {isOpen && filteredCountries.length > 0 && (
        <div className="absolute z-50 w-full mt-1 p-2 shadow-mid bg-level-1 overflow-y-auto overflow-x-hidden rounded-2xl border-border-low max-h-60">
          {filteredCountries.map((option: { value: string; label: string }) => (
            <button
              key={option.value}
              type="button"
              onClick={() => handleCountryClick(option.value)}
              className="w-full px-3 py-2 text-left text-sm text-txt-primary hover:bg-highlight-hover focus:bg-highlight-pressed focus:outline-none rounded-lg"
            >
              {option.label}
            </button>
          ))}
        </div>
      )}

      {isOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setIsOpen(false)}
        />
      )}
    </div>
  );
}
