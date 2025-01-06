export const SumSVG = (props: React.HTMLAttributes<SVGElement>) => (
  <svg
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <g id="sum">
      <g id="Icon">
        <path
          d="M0 0h24v24H0z"
          stroke="none"
        />
        <path
          d="M18 16v2a1 1 0 01-1 1H6l6-7-6-7h11a1 1 0 011 1v2"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </g>
    </g>
  </svg>
);
