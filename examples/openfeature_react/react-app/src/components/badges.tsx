import {OpenFeature, useFlag} from "@openfeature/react-sdk";

export const Badges = () => {
  const {value: badgeClass} = useFlag("badge-class", "");
  const userType = OpenFeature.getContext().userType ?? "" as string;

  return <div className="flex justify-center items-center">
    {badgeClass && <span className={badgeClass}>{userType.toString()}</span>}
  </div>
}