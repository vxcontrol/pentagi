export const ReportSVG = (props: React.HTMLAttributes<SVGElement>) => (
  <svg
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <g id="report">
      <g id="Icon">
        <path
          d="M0 0h24v24H0z"
          stroke="none"
        />
        <path
          d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2"
          stroke="currentColor"
          strokeWidth={1.7}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M11 3 H13 A2 2 0 0 1 15 5 V5 A2 2 0 0 1 13 7 H11 A2 2 0 0 1 9 5 V5 A2 2 0 0 1 11 3 z"
          stroke="currentColor"
          strokeWidth={1.7}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M9 17v-5M12 17v-1M15 17v-3"
          stroke="currentColor"
          strokeWidth={1.7}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </g>
    </g>
  </svg>
);
