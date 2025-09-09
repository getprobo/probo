import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  Button,
  Checkbox,
  IconLock,
  IconArrowDown,
  useToast,
  useDialogRef
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useAcceptNonDisclosureAgreement } from "/hooks/useTrustCenterQueries";
import { buildEndpoint } from "/providers/RelayProviders";
import { sprintf } from "@probo/helpers";

type Props = {
  trustCenterId: string;
  organizationName: string;
  ndaFileName?: string | null;
  ndaFileUrl?: string | null;
};

export function NDAAcceptanceDialog({ trustCenterId, organizationName, ndaFileName, ndaFileUrl }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [isChecked, setIsChecked] = useState(false);
  const dialogRef = useDialogRef();
  const acceptNdaMutation = useAcceptNonDisclosureAgreement();

  useEffect(() => {
    dialogRef.current?.open();
  }, []);

  const handleLogout = async () => {
    try {
      const response = await fetch(buildEndpoint('/api/trust/v1/auth/logout'), {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error("Logout failed");
      }

      window.location.reload();
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Logout failed"),
        variant: "error",
      });
    }
  };

  const handleAccept = () => {
    if (!isChecked) {
      toast({
        title: __("Agreement Required"),
        description: __("Please check the box to confirm your agreement"),
        variant: "error",
      });
      return;
    }

    acceptNdaMutation.mutate(
      { trustCenterId },
      {
        onSuccess: () => {
          window.location.reload();
        },
        onError: () => {
          toast({
            title: __("Error"),
            description: __("Failed to accept the Non-Disclosure Agreement"),
            variant: "error",
          });
        },
      }
    );
  };

  const handleCancel = () => {
    handleLogout();
  };

  return (
    <Dialog
      ref={dialogRef}
      closable={false}
      onClose={handleCancel}
    >
        <DialogContent>
        <div className="space-y-3 p-4">
          <div className="text-center">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-amber-50 border border-amber-200 mb-3">
              <IconLock className="h-6 w-6 text-amber-600" />
            </div>
            <h2 className="text-lg font-semibold text-txt-primary mb-2">
              {__("Non-Disclosure Agreement")}
            </h2>
            <p className="text-sm text-txt-secondary">
              {sprintf(__("To access %s's trust center, you must accept the Non-Disclosure Agreement."), organizationName)}
            </p>
          </div>

          <div className="bg-level-1 p-3 rounded-lg border border-border-subtle">
            {ndaFileName && ndaFileUrl ? (
              <div className="text-center">
                <p className="text-sm font-medium text-txt-primary mb-3">
                  {__("Please review and download the Non-Disclosure Agreement:")}
                </p>
                <div className="flex justify-center">
                  <Button
                    variant="secondary"
                    icon={IconArrowDown}
                    onClick={() => {
                      const link = document.createElement('a');
                      link.href = ndaFileUrl;
                      link.download = ndaFileName || 'NDA.pdf';
                      link.target = '_blank';
                      link.rel = 'noopener noreferrer';
                      document.body.appendChild(link);
                      link.click();
                      document.body.removeChild(link);
                    }}
                  >
                    {sprintf(__("Download %s"), ndaFileName)}
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <p className="text-sm font-medium text-txt-primary mb-2">
                  {__("By accepting this agreement, you commit to:")}
                </p>
                <ul className="text-sm text-txt-secondary space-y-0.5">
                  <li className="flex items-start">
                    <span className="inline-block w-2 h-2 rounded-full bg-txt-tertiary mt-1.5 mr-2 flex-shrink-0"></span>
                    {__("Keep confidential information secure")}
                  </li>
                  <li className="flex items-start">
                    <span className="inline-block w-2 h-2 rounded-full bg-txt-tertiary mt-1.5 mr-2 flex-shrink-0"></span>
                    {__("Not share or disclose sensitive data")}
                  </li>
                  <li className="flex items-start">
                    <span className="inline-block w-2 h-2 rounded-full bg-txt-tertiary mt-1.5 mr-2 flex-shrink-0"></span>
                    {__("Use information only for authorized purposes")}
                  </li>
                </ul>
              </>
            )}
          </div>

          <div className="flex items-start space-x-2 p-3 bg-level-0 rounded-lg border border-border-subtle">
            <div className="mt-0.5">
              <Checkbox
                checked={isChecked}
                onChange={setIsChecked}
              />
            </div>
            <label
              className="text-sm text-txt-primary cursor-pointer flex-1"
              onClick={() => setIsChecked(!isChecked)}
            >
              {__("I agree to the terms of the Non-Disclosure Agreement and will handle all information accordingly.")}
            </label>
          </div>
        </div>
      </DialogContent>

      <DialogFooter exitLabel={__("Disconnect")}>
        <Button
          variant="primary"
          onClick={handleAccept}
          disabled={!isChecked || acceptNdaMutation.isPending}
        >
          {acceptNdaMutation.isPending ? __("Accepting...") : __("Accept & Continue")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
