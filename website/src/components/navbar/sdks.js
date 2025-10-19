import {sdk} from '../../../data/sdk.js';

// Function to generate HTML string for SDKs dropdown
export function generateSdksDropdownHTML() {
  const clientSdks = sdk.filter(s => s.paradigm.includes('Client'));
  const serverSdks = sdk.filter(s => s.paradigm.includes('Server'));

  const generateSdkSection = sectionItem => {
    return `
      <section class="flex flex-col gap-3 mb-5">
      <div class="text-center lg:text-left">
        <h3 class="text-2xl font-poppins font-[800] tracking-[-0.08rem] mb-0">
          ${sectionItem.title}
        </h3>
        <p class="text-xs text-gray-400 font-poppins font-[400] leading-8 mb-0">
          ${sectionItem.description}
        </p>
      </div>

      <div class="inline-flex flex-wrap gap-2">
        ${sectionItem.sdks.map(sdkItem => generateSdkCard(sdkItem)).join('')}
      </div>
    </section>
    `;
  };

  const generateSdkCard = sdkItem => {
    return `
        <a
          href="/docs/sdk/${sdkItem.docLink}"
          class="no-underline hover:no-underline inline-flex items-center gap-x-1.5 py-1.5 px-3 rounded-lg text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-800/30 dark:text-blue-500"
        >
          <i class="${sdkItem.faLogo} text-xl"></i>
          ${sdkItem.name}
        </a>
    `;
  };

  return `
    <div class="sdks-dropdown flex max-w-4xl flex-col rounded-2xl lg:min-w-[600px]">
      <div class="flex flex-col bg-secondary-800lg:min-h-[300px]">
        ${generateSdkSection({
          title: 'Client SDKs',
          description:
            'Use feature flags in your web or mobile applications with these SDKs.',
          sdks: clientSdks,
        })}  
        ${generateSdkSection({
          title: 'Server SDKs',
          description:
            'Use feature flags in your backend applications with these SDKs.',
          sdks: serverSdks,
        })}
    </div>
  `;
}
