export function singleFlagFormConvertor(flagFormData) {
  const variationType = flagFormData.type;
  const variations = {};

  flagFormData.variations
    .filter(
      i =>
        i.name !== undefined &&
        i.name !== '' &&
        i.value !== undefined &&
        i.value !== ''
    )
    .forEach(
      i => (variations[i.name] = convertValueIntoType(i.value, variationType))
    );

  const targeting = flagFormData.targeting.map(t => convertRule(t));
  const trackEvents = convertValueIntoType(flagFormData.trackEvents, 'boolean');
  const disable = convertValueIntoType(flagFormData.disable, 'boolean');
  const defaultRule = convertRule(flagFormData.defaultRule);

  return {
    variations,
    disable: !disable ? undefined : disable,
    trackEvents: trackEvents ? undefined : trackEvents,
    version: flagFormData.version === '' ? undefined : flagFormData.version,
    targeting: targeting.length > 0 ? targeting : undefined,
    defaultRule,
    metadata: convertMetadata(flagFormData.metadata),
  };
}

function convertValueIntoType(value, type) {
  switch (type) {
    case 'json':
      try {
        return JSON.parse(value.value);
      } catch (e) {
        // Invalid JSON, return undefined
        return undefined;
      }
    case 'number':
      return Number(value) || undefined;
    case 'boolean':
      if (typeof value == 'boolean') return value;
      return (
        value !== undefined &&
        (typeof value === 'string' || value instanceof String) &&
        value.toLowerCase() === 'true'
      );
    default:
      return String(value) || undefined;
  }
}

function convertMetadata(metadata) {
  if (
    metadata === undefined ||
    metadata.filter(({name}) => name !== '').length === 0
  ) {
    return undefined;
  }
  return Object.assign(
    {},
    ...metadata.map(({name, value}) => ({[name]: value}))
  );
}

function convertRule(ruleForm) {
  let variation,
    percentage,
    progressiveRollout;
  const {selectedVar} = ruleForm;
  switch (selectedVar) {
    case 'percentage':
      percentage = {};
      ruleForm.percentages.forEach(i => (percentage[i.name] = i.value));
      break;
    case 'progressive':
      progressiveRollout = {
        initial: {
          variation: ruleForm.progressive.initial.selectedVar,
          percentage: ruleForm.progressive.initial.percentage || 0,
          date: ruleForm.progressive.initial.date,
        },
        end: {
          variation: ruleForm.progressive.end.selectedVar,
          percentage: ruleForm.progressive.end.percentage || 100,
          date: ruleForm.progressive.end.date,
        },
      };
      break;
    default:
      variation = selectedVar;
      break;
  }

  return {
    name: ruleForm.name || undefined,
    query: ruleForm.query,
    variation,
    percentage,
    progressiveRollout,
  };
}
