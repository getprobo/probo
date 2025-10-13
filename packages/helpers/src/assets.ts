export const getAssetTypeVariant = (type: string) => {
  switch (type) {
    case "PHYSICAL":
      return "neutral";
    case "VIRTUAL":
      return "info";
    default:
      return "neutral";
  }
};
