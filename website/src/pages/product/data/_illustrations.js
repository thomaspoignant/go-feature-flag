import React from 'react';
import PropTypes from 'prop-types';

// Co-located inline SVG illustrations for the /product/data landing page.
// Each one is a self-contained framed panel using the GO Feature Flag teal
// (`goff`) palette, so it reads well on both the light and dark theme.

const WRAP = 'mx-auto block h-auto w-full max-w-xl rounded-2xl md:max-w-none';

// Shared colors (Tailwind `goff` scale).
const INK = '#0a4a41'; // goff-900 - primary label text
const INK_SOFT = '#0a7263'; // goff-700 - secondary text
const CARD = '#ffffff';
const CARD_STROKE = '#abefd9'; // goff-200
const ACCENT = '#18b192'; // goff-500
const ACCENT_DEEP = '#0c8f77'; // goff-600
const STREAM = '#74e1c4'; // goff-300

function FlagGlyph({x, y}) {
  return (
    <g transform={`translate(${x} ${y})`}>
      <line
        x1="0"
        y1="-26"
        x2="0"
        y2="34"
        stroke={INK_SOFT}
        strokeWidth="4"
        strokeLinecap="round"
      />
      <path d="M0 -22 L46 -12 L30 2 L46 16 L0 24 Z" fill={ACCENT} />
    </g>
  );
}

FlagGlyph.propTypes = {
  x: PropTypes.number.isRequired,
  y: PropTypes.number.isRequired,
};

/* 1. The core idea: evaluations + outcomes routed (never computed) out to
   many destinations. */
export function CoreIdeaIllustration() {
  return (
    <svg
      viewBox="0 0 1200 655"
      role="img"
      aria-label="GO Feature Flag routing flag evaluations and outcomes out to a data warehouse, a queue, and a data lake."
      className={WRAP}
      fontFamily="inherit">
      <defs>
        <linearGradient id="dataPanelA" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#f4fdfa" />
          <stop offset="1" stopColor="#e2f6ee" />
        </linearGradient>
      </defs>
      <rect
        x="1"
        y="1"
        width="1198"
        height="653"
        rx="24"
        fill="url(#dataPanelA)"
        stroke="#cdf7e7"
        strokeWidth="2"
      />

      {/* connectors fanning out from the hub */}
      <g fill="none" stroke={STREAM} strokeWidth="3">
        <path d="M320 328 L430 328" />
        <path d="M660 328 C 780 328, 800 145, 900 145" />
        <path d="M660 328 C 780 328, 800 327, 900 327" />
        <path d="M660 328 C 780 328, 800 509, 900 509" />
      </g>
      {/* event dots travelling the streams */}
      <g fill={ACCENT}>
        <circle cx="372" cy="328" r="6" />
        <circle cx="742" cy="223" r="6" />
        <circle cx="800" cy="327" r="6" />
        <circle cx="742" cy="433" r="6" />
      </g>

      {/* source card */}
      <rect
        x="70"
        y="250"
        width="250"
        height="156"
        rx="16"
        fill={CARD}
        stroke={CARD_STROKE}
        strokeWidth="2"
      />
      <FlagGlyph x={155} y={312} />
      <text
        x="195"
        y="372"
        textAnchor="middle"
        fontSize="22"
        fontWeight="700"
        fill={INK}>
        Evaluations
      </text>
      <text x="195" y="394" textAnchor="middle" fontSize="15" fill={INK_SOFT}>
        &amp; Track() outcomes
      </text>

      {/* router hub */}
      <rect
        x="430"
        y="250"
        width="230"
        height="156"
        rx="24"
        fill={ACCENT_DEEP}
      />
      <text
        x="545"
        y="318"
        textAnchor="middle"
        fontSize="24"
        fontWeight="700"
        fill="#ffffff">
        GO Feature Flag
      </text>
      <text x="545" y="348" textAnchor="middle" fontSize="16" fill="#cdf7e7">
        routes &mdash; never computes
      </text>

      {/* destination cards */}
      <g>
        {/* warehouse */}
        <rect
          x="900"
          y="70"
          width="230"
          height="150"
          rx="16"
          fill={CARD}
          stroke={CARD_STROKE}
          strokeWidth="2"
        />
        <path d="M928 158 L958 130 L988 158 Z" fill={ACCENT} />
        <rect x="936" y="158" width="44" height="34" rx="3" fill="#cdf7e7" />
        <rect x="952" y="170" width="12" height="22" fill={ACCENT_DEEP} />
        <text x="1010" y="140" fontSize="20" fontWeight="700" fill={INK}>
          Warehouse
        </text>
        <text x="1010" y="164" fontSize="14" fill={INK_SOFT}>
          BigQuery, S3&hellip;
        </text>

        {/* queue / stream */}
        <rect
          x="900"
          y="252"
          width="230"
          height="150"
          rx="16"
          fill={CARD}
          stroke={CARD_STROKE}
          strokeWidth="2"
        />
        <rect x="930" y="300" width="56" height="14" rx="4" fill={STREAM} />
        <rect x="930" y="320" width="56" height="14" rx="4" fill={ACCENT} />
        <rect
          x="930"
          y="340"
          width="56"
          height="14"
          rx="4"
          fill={ACCENT_DEEP}
        />
        <text x="1010" y="322" fontSize="20" fontWeight="700" fill={INK}>
          Queue
        </text>
        <text x="1010" y="346" fontSize="14" fill={INK_SOFT}>
          Kafka, Pub/Sub&hellip;
        </text>

        {/* data lake */}
        <rect
          x="900"
          y="434"
          width="230"
          height="150"
          rx="16"
          fill={CARD}
          stroke={CARD_STROKE}
          strokeWidth="2"
        />
        <g fill={ACCENT} stroke={ACCENT_DEEP} strokeWidth="2">
          <path d="M932 492 v34 a26 9 0 0 0 52 0 v-34" fill="#cdf7e7" />
          <ellipse cx="958" cy="492" rx="26" ry="9" />
        </g>
        <text x="1010" y="504" fontSize="20" fontWeight="700" fill={INK}>
          Data lake
        </text>
        <text x="1010" y="528" fontSize="14" fill={INK_SOFT}>
          GCS, Blob&hellip;
        </text>
      </g>
    </svg>
  );
}

/* 2. Pattern 1: the provider and relay proxy emit one "feature" event per
   evaluation, automatically. */
export function ProviderEventsIllustration() {
  return (
    <svg
      viewBox="0 0 1200 896"
      role="img"
      aria-label="An OpenFeature provider emitting a feature event per evaluation through the relay proxy to an exporter."
      className={WRAP}
      fontFamily="inherit">
      <defs>
        <linearGradient id="dataPanelB" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#f4fdfa" />
          <stop offset="1" stopColor="#e2f6ee" />
        </linearGradient>
      </defs>
      <rect
        x="1"
        y="1"
        width="1198"
        height="894"
        rx="24"
        fill="url(#dataPanelB)"
        stroke="#cdf7e7"
        strokeWidth="2"
      />

      <text
        x="600"
        y="150"
        textAnchor="middle"
        fontSize="26"
        fontWeight="700"
        fill={INK}>
        Built in &mdash; no application code
      </text>

      {/* flow connectors */}
      <g fill="none" stroke={STREAM} strokeWidth="3">
        <path d="M392 470 L520 470" />
        <path d="M760 470 L880 470" />
      </g>
      <g fill={ACCENT}>
        <circle cx="456" cy="470" r="6" />
        <circle cx="820" cy="470" r="6" />
      </g>

      {/* provider card with evaluations */}
      <rect
        x="72"
        y="320"
        width="320"
        height="300"
        rx="20"
        fill={CARD}
        stroke={CARD_STROKE}
        strokeWidth="2"
      />
      <text
        x="232"
        y="362"
        textAnchor="middle"
        fontSize="20"
        fontWeight="700"
        fill={INK}>
        OpenFeature provider
      </text>
      {[0, 1, 2].map(i => (
        <g key={i} transform={`translate(0 ${i * 58})`}>
          <rect
            x="100"
            y="392"
            width="264"
            height="44"
            rx="12"
            fill="#edfcf7"
            stroke={CARD_STROKE}
            strokeWidth="2"
          />
          <circle cx="124" cy="414" r="11" fill={ACCENT} />
          <path
            d="M118 414 l5 5 l8 -10"
            fill="none"
            stroke="#ffffff"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          <text x="146" y="420" fontSize="16" fill={INK_SOFT}>
            flag evaluated
          </text>
        </g>
      ))}

      {/* relay proxy card */}
      <rect
        x="520"
        y="384"
        width="240"
        height="172"
        rx="20"
        fill={ACCENT_DEEP}
      />
      <text
        x="640"
        y="458"
        textAnchor="middle"
        fontSize="22"
        fontWeight="700"
        fill="#ffffff">
        Relay proxy
      </text>
      <text x="640" y="486" textAnchor="middle" fontSize="15" fill="#cdf7e7">
        data collector
      </text>

      {/* exporter card */}
      <rect
        x="880"
        y="384"
        width="248"
        height="172"
        rx="20"
        fill={CARD}
        stroke={CARD_STROKE}
        strokeWidth="2"
      />
      <text
        x="1004"
        y="450"
        textAnchor="middle"
        fontSize="22"
        fontWeight="700"
        fill={INK}>
        Exporter
      </text>
      <g fill={STREAM}>
        <rect x="936" y="476" width="40" height="14" rx="4" />
        <rect x="984" y="476" width="40" height="14" rx="4" fill={ACCENT} />
        <rect
          x="1032"
          y="476"
          width="40"
          height="14"
          rx="4"
          fill={ACCENT_DEEP}
        />
      </g>

      {/* "feature" event chips on the path */}
      <g>
        <rect
          x="410"
          y="640"
          width="120"
          height="40"
          rx="20"
          fill="#cdf7e7"
          stroke={STREAM}
          strokeWidth="2"
        />
        <text
          x="470"
          y="666"
          textAnchor="middle"
          fontSize="16"
          fontWeight="700"
          fill={INK}>
          feature
        </text>
        <rect
          x="678"
          y="640"
          width="120"
          height="40"
          rx="20"
          fill="#cdf7e7"
          stroke={STREAM}
          strokeWidth="2"
        />
        <text
          x="738"
          y="666"
          textAnchor="middle"
          fontSize="16"
          fontWeight="700"
          fill={INK}>
          feature
        </text>
      </g>

      <text
        x="600"
        y="752"
        textAnchor="middle"
        fontSize="20"
        fontWeight="600"
        fill={INK_SOFT}>
        Every evaluation &rarr; one feature event, shipped automatically
      </text>
    </svg>
  );
}

/* 3. Pattern 2: the Track API records an outcome that joins back to the flag a
   user saw, via the shared context key. */
export function TrackJoinIllustration() {
  return (
    <svg
      viewBox="0 0 1200 896"
      role="img"
      aria-label="An exposure event and a Track() outcome sharing the same context key, joined to measure the winning variation."
      className={WRAP}
      fontFamily="inherit">
      <defs>
        <linearGradient id="dataPanelC" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#f4fdfa" />
          <stop offset="1" stopColor="#e2f6ee" />
        </linearGradient>
      </defs>
      <rect
        x="1"
        y="1"
        width="1198"
        height="894"
        rx="24"
        fill="url(#dataPanelC)"
        stroke="#cdf7e7"
        strokeWidth="2"
      />

      {/* dashed links converging on the shared key */}
      <g fill="none" stroke="#3ccbaa" strokeWidth="3" strokeDasharray="2 9">
        <path d="M560 320 C 640 320, 600 448, 640 448" />
        <path d="M560 628 C 640 628, 600 448, 640 448" />
      </g>
      {/* arrow key -> result */}
      <g fill="none" stroke={STREAM} strokeWidth="3">
        <path d="M712 448 L760 448" />
      </g>

      {/* exposure card */}
      <rect
        x="120"
        y="230"
        width="440"
        height="180"
        rx="20"
        fill={CARD}
        stroke={CARD_STROKE}
        strokeWidth="2"
      />
      <text
        x="150"
        y="276"
        fontSize="15"
        fontWeight="700"
        letterSpacing="1.5"
        fill={ACCENT_DEEP}>
        EXPOSURE
      </text>
      <text x="150" y="320" fontSize="22" fontWeight="700" fill={INK}>
        new-checkout
      </text>
      <rect x="150" y="344" width="86" height="36" rx="10" fill={ACCENT} />
      <text
        x="193"
        y="369"
        textAnchor="middle"
        fontSize="17"
        fontWeight="700"
        fill="#ffffff">
        var B
      </text>
      {/* key chip */}
      <g>
        <rect
          x="404"
          y="344"
          width="128"
          height="36"
          rx="18"
          fill="#edfcf7"
          stroke={STREAM}
          strokeWidth="2"
        />
        <circle
          cx="426"
          cy="362"
          r="7"
          fill="none"
          stroke={INK_SOFT}
          strokeWidth="3"
        />
        <line
          x1="431"
          y1="362"
          x2="448"
          y2="362"
          stroke={INK_SOFT}
          strokeWidth="3"
          strokeLinecap="round"
        />
        <text x="458" y="368" fontSize="15" fontWeight="600" fill={INK_SOFT}>
          u_8f3
        </text>
      </g>

      {/* outcome card */}
      <rect
        x="120"
        y="540"
        width="440"
        height="180"
        rx="20"
        fill={CARD}
        stroke={CARD_STROKE}
        strokeWidth="2"
      />
      <text
        x="150"
        y="586"
        fontSize="15"
        fontWeight="700"
        letterSpacing="1.5"
        fill={ACCENT_DEEP}>
        OUTCOME &middot; Track() API
      </text>
      <text x="150" y="630" fontSize="22" fontWeight="700" fill={INK}>
        checkout-completed
      </text>
      <rect x="150" y="654" width="104" height="36" rx="10" fill="#cdf7e7" />
      <text
        x="202"
        y="679"
        textAnchor="middle"
        fontSize="17"
        fontWeight="700"
        fill={INK}>
        $99.99
      </text>
      {/* key chip (same) */}
      <g>
        <rect
          x="404"
          y="654"
          width="128"
          height="36"
          rx="18"
          fill="#edfcf7"
          stroke={STREAM}
          strokeWidth="2"
        />
        <circle
          cx="426"
          cy="672"
          r="7"
          fill="none"
          stroke={INK_SOFT}
          strokeWidth="3"
        />
        <line
          x1="431"
          y1="672"
          x2="448"
          y2="672"
          stroke={INK_SOFT}
          strokeWidth="3"
          strokeLinecap="round"
        />
        <text x="458" y="678" fontSize="15" fontWeight="600" fill={INK_SOFT}>
          u_8f3
        </text>
      </g>

      {/* shared key node */}
      <circle cx="676" cy="448" r="36" fill={ACCENT_DEEP} />
      <circle
        cx="668"
        cy="442"
        r="9"
        fill="none"
        stroke="#ffffff"
        strokeWidth="4"
      />
      <line
        x1="675"
        y1="449"
        x2="692"
        y2="466"
        stroke="#ffffff"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <text
        x="676"
        y="524"
        textAnchor="middle"
        fontSize="16"
        fontWeight="600"
        fill={INK_SOFT}>
        same context key
      </text>

      {/* result card */}
      <rect
        x="760"
        y="356"
        width="320"
        height="184"
        rx="20"
        fill="#edfcf7"
        stroke={STREAM}
        strokeWidth="2"
      />
      <text
        x="920"
        y="404"
        textAnchor="middle"
        fontSize="15"
        fontWeight="700"
        letterSpacing="1.5"
        fill={ACCENT_DEEP}>
        MEASURED
      </text>
      <text
        x="920"
        y="446"
        textAnchor="middle"
        fontSize="22"
        fontWeight="700"
        fill={INK}>
        Variation B converted
      </text>
      {/* mini bars */}
      <g>
        <rect x="838" y="476" width="64" height="20" rx="5" fill={STREAM} />
        <rect x="838" y="502" width="150" height="20" rx="5" fill={ACCENT} />
        <text x="912" y="517" fontSize="13" fontWeight="700" fill="#ffffff">
          winner
        </text>
      </g>
    </svg>
  );
}
