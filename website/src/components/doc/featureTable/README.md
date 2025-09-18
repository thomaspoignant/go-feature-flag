# FeatureTable Component

A reusable React component that displays a table of features for a given SDK, showing the status of each feature with appropriate icons and styling.

## Usage

```jsx
import FeatureTable from '@site/src/components/doc/featureTable';

const MyComponent = () => {
  const sdk = {
    name: 'Kotlin / Android',
    featureList: [
      {
        name: 'Flag evaluation',
        status: 'done',
        description: 'It is possible to evaluate all the type of flags'
      },
      {
        name: 'Cache invalidation',
        status: 'done',
        description: 'A polling mechanism is in place to refresh the cache'
      },
      {
        name: 'Logging',
        status: 'not implemented',
        description: 'Not supported by the SDK'
      }
    ]
  };

  return <FeatureTable sdk={sdk} />;
};
```

## Props

### `sdk` (required)
An object containing the SDK information:

- `name` (string): The name of the SDK
- `featureList` (array): Array of feature objects

Each feature object should have:
- `name` (string, required): The name of the feature
- `status` (string, required): The status of the feature
- `description` (string, optional): A description of the feature

## Supported Status Values

The component recognizes the following status values (case-insensitive):

- `'done'` or `'implemented'` → ✅ Green checkmark
- `'not implemented'` or `'not implemented yet'` → ❌ Red X
- `'in progress'` or `'in-progress'` → ⚠️ Yellow clock
- Any other value → ❓ Gray question mark

## Features

- **Responsive Design**: Table adapts to different screen sizes
- **Dark Mode Support**: Compatible with dark/light theme switching
- **Hover Effects**: Rows highlight on hover
- **Accessibility**: Proper table structure with semantic HTML
- **Icon Consistency**: Uses FontAwesome icons matching the existing design system
- **TypeScript Support**: Includes PropTypes for runtime type checking

## Styling

The component uses Tailwind CSS classes and follows the existing design patterns in the project. It includes:

- Border styling for table cells
- Hover effects on table rows
- Dark mode support
- Consistent spacing and typography
- Status legend at the bottom

## Example Output

The component generates a table similar to the one in the Android documentation, with:

1. A header row with "Status", "Feature", and "Description" columns
2. Feature rows with status icons, feature names, and descriptions
3. A legend explaining the status icons 