import {sdk} from '../../../data/sdk.js';

// Function to generate HTML string for SDKs dropdown
export function generateSdksDropdownHTML() {
  const clientSdks = sdk.filter(s => s.paradigm.includes('Client'));
  const serverSdks = sdk.filter(s => s.paradigm.includes('Server'));

  const generateSdkBadge = sdkItem => `
    <a
      href="/${sdkItem.docLink}"
      class="no-underline hover:no-underline inline-flex items-center gap-x-1.5 py-1.5 px-3 rounded-lg text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-800/30 dark:text-blue-500"
    >
      <i class="${sdkItem.faLogo} text-xl"></i>
      ${sdkItem.name}
    </a>
  `;

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

// Function to create and return a DOM element for the SDKs dropdown
export function createSdksDropdownElement() {
  const container = document.createElement('div');
  container.innerHTML = generateSdksDropdownHTML();
  return container.firstElementChild;
}

// Function to render the SDKs dropdown into a specific DOM element
export function renderSdksDropdown(containerId) {
  const container = document.getElementById(containerId);
  if (container) {
    container.innerHTML = generateSdksDropdownHTML();
  } else {
    console.error(`Container with id "${containerId}" not found`);
  }
}

// Function to append the SDKs dropdown to a specific DOM element
export function appendSdksDropdown(containerId) {
  const container = document.getElementById(containerId);
  if (container) {
    const dropdownElement = createSdksDropdownElement();
    container.appendChild(dropdownElement);
  } else {
    console.error(`Container with id "${containerId}" not found`);
  }
}

// Function to get client SDKs data
export function getClientSdks() {
  return sdk.filter(s => s.paradigm.includes('Client'));
}

// Function to get server SDKs data
export function getServerSdks() {
  return sdk.filter(s => s.paradigm.includes('Server'));
}

// Function to get SDK by key
export function getSdkByKey(key) {
  return sdk.find(s => s.key === key);
}

// Function to get SDKs by paradigm
export function getSdksByParadigm(paradigm) {
  return sdk.filter(s => s.paradigm.includes(paradigm));
}

// Utility function to create SDK badge element
export function createSdkBadgeElement(sdkItem) {
  const link = document.createElement('a');
  link.href = `/${sdkItem.docLink}`;
  link.className =
    'no-underline hover:no-underline inline-flex items-center gap-x-1.5 py-1.5 px-3 rounded-lg text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-800/30 dark:text-blue-500';

  const icon = document.createElement('i');
  icon.className = `${sdkItem.faLogo} text-xl`;

  link.appendChild(icon);
  link.appendChild(document.createTextNode(sdkItem.name));

  return link;
}

// Utility function to create SDK card element
export function createSdkCardElement(sdkItem) {
  const colorClasses = {
    'devicon-javascript-plain':
      'bg-yellow-100 dark:bg-yellow-900 hover:border-yellow-300 dark:hover:border-yellow-600',
    'devicon-typescript-plain':
      'bg-blue-100 dark:bg-blue-900 hover:border-blue-300 dark:hover:border-blue-600',
    'devicon-react-original':
      'bg-cyan-100 dark:bg-cyan-900 hover:border-cyan-300 dark:hover:border-cyan-600',
    'devicon-angularjs-plain':
      'bg-red-100 dark:bg-red-900 hover:border-red-300 dark:hover:border-red-600',
    'devicon-swift-plain':
      'bg-orange-100 dark:bg-orange-900 hover:border-orange-300 dark:hover:border-orange-600',
    'devicon-android-plain':
      'bg-green-100 dark:bg-green-900 hover:border-green-300 dark:hover:border-green-600',
  };

  const colorClass =
    colorClasses[sdkItem.faLogo] ||
    'bg-gray-100 dark:bg-gray-900 hover:border-gray-300 dark:hover:border-gray-600';

  const link = document.createElement('a');
  link.href = `/${sdkItem.docLink}`;
  link.className = `group flex items-center gap-4 rounded-lg bg-white dark:bg-gray-800 shadow-sm border border-gray-200 dark:border-gray-700 transition-all hover:shadow-md ${colorClass}`;

  const iconContainer = document.createElement('div');
  iconContainer.className = `flex-shrink-0 w-10 h-10 rounded-full ${
    colorClass.split(' ')[0]
  } flex items-center justify-center`;

  const icon = document.createElement('i');
  icon.className = `${sdkItem.faLogo} text-xl`;

  const span = document.createElement('span');
  span.className = 'text-sm font-medium text-gray-900 dark:text-white';
  span.textContent = sdkItem.name;

  iconContainer.appendChild(icon);
  link.appendChild(iconContainer);
  link.appendChild(span);

  return link;
}
