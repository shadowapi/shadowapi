import { FrontendApi } from "@ory/client";
import Configuration from "./config";
import { NavigateFunction } from "react-router-dom"
import { UseFormReset, FieldValues } from "react-hook-form"

export const ory = new FrontendApi(Configuration)

export function handleFlowError<FormFields extends FieldValues>(
  navigate: NavigateFunction,
  flowType: "login" | "signup" | "settings" | "recovery" | "verification",
  resetForm: UseFormReset<FormFields>,
) {
  return async (err: any) => {
    switch (err.response?.data.error?.id) {
      case "session_inactive":
        navigate("/login?return_to=" + window.location.href);
        console.log("session_inactive")
        return;
      case "session_aal2_required":
        if (err.response?.data.redirect_browser_to) {
          const redirectTo = new URL(err.response?.data.redirect_browser_to)
          if (flowType === "settings") {
            redirectTo.searchParams.set("return_to", window.location.href)
          }
          window.location.href = redirectTo.toString();
          return;
        }
        navigate("/login?aal=aal2&return_to=" + window.location.href)
        console.log("aal2_required")
        return
      case "session_already_available":
        navigate("/");
        console.log("session_already_available")
        return;
      case "session_refresh_required":
        window.location.href = err.response?.data.redirect_browser_to;
        console.log("session_refresh_required")
        return;
      case "self_service_flow_return_to_forbidden":
        resetForm();
        navigate("/" + flowType);
        console.log("self_service_flow_return_to_forbidden")
        return;
      case "self_service_flow_expired":
        resetForm();
        navigate("/" + flowType);
        console.log("self_service_flow_expired")
        return;
      case "security_csrf_violation":
        resetForm();
        navigate("/" + flowType);
        console.log("security_csrf_violation")
        return;
      case "security_identity_mismatch":
        navigate("/" + flowType);
        console.log("security_identity_mismatch")
        return;
      case "browser_location_change_required":
        window.location.href = err.response.data.redirect_browser_to;
        console.log("browser_location_change_required")
        return;
    }

    switch (err.response?.status) {
      case 410:
        resetForm();
        navigate("/" + flowType);
        console.log("410")
        return;
    }

    return Promise.reject(err);
  };
}
