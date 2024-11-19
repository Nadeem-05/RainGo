<script lang="ts">
  import { onMount } from "svelte";
  import Chart from "chart.js/auto";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    NavHamburger,
  } from "flowbite-svelte";
  const chartConfigs = [
    {
      id: "lineChart",
      type: "line",
      label: "Acquisitions Over Time",
      data: [
        { year: 2010, count: 10 },
        { year: 2011, count: 20 },
        { year: 2012, count: 15 },
        { year: 2013, count: 25 },
        { year: 2014, count: 22 },
        { year: 2015, count: 30 },
        { year: 2016, count: 28 },
      ],
    },
    {
      id: "barChart",
      type: "bar",
      label: "Sales by Year",
      data: [
        { year: 2010, sales: 200 },
        { year: 2011, sales: 250 },
        { year: 2012, sales: 300 },
        { year: 2013, sales: 280 },
        { year: 2014, sales: 320 },
      ],
    },
    {
      id: "pieChart",
      type: "pie",
      label: "Product Distribution",
      data: [
        { category: "Product A", count: 50 },
        { category: "Product B", count: 30 },
        { category: "Product C", count: 20 },
      ],
    },
    {
      id: "doughnutChart",
      type: "doughnut",
      label: "Market Share",
      data: [
        { region: "North America", share: 35 },
        { region: "Europe", share: 40 },
        { region: "Asia", share: 25 },
      ],
    },
  ];

  onMount(() => {
    chartConfigs.forEach((config) => {
      const ctx = document.getElementById(config.id) as HTMLCanvasElement;
      new Chart(ctx, {
        type: config.type,
        data: {
          labels: config.data.map(
            (row) => row.year || row.category || row.region,
          ),
          datasets: [
            {
              label: config.label,
              data: config.data.map(
                (row) => row.count || row.sales || row.share,
              ),
              backgroundColor: [
                "rgba(54, 162, 235, 0.2)",
                "rgba(255, 99, 132, 0.2)",
                "rgba(75, 192, 192, 0.2)",
              ],
              borderColor: [
                "rgba(54, 162, 235, 1)",
                "rgba(255, 99, 132, 1)",
                "rgba(75, 192, 192, 1)",
              ],
              borderWidth: 2,
            },
          ],
        },
        options: {
          responsive: true,
          plugins: {
            legend: {
              display: true,
              position: "top",
              labels: { color: "#fff" },
            },
          },
          scales: {
            x: { ticks: { color: "#fff" } },
            y: { ticks: { color: "#fff" }, beginAtZero: true },
          },
        },
      });
    });
  });
</script>

<Navbar rounded class="bg-transparent dark">
  <NavBrand href="/">
    <img
      src="src/images/favicon.png"
      class="me-3 h-6 sm:h-9"
      alt="Flowbite Logo"
    />
    <span class="self-center text-xl font-semibold dark:text-white"
      >Flowbite</span
    >
  </NavBrand>
  <NavHamburger />
  <NavUl>
    <NavLi href="/" activeClass="active" class="dark:text-white">Home</NavLi>
    <NavLi href="/about" class="dark:text-white">About</NavLi>
    <NavLi href="/docs/components/navbar" class="dark:text-white">Navbar</NavLi>
    <NavLi href="/pricing" class="dark:text-white">Pricing</NavLi>
    <NavLi href="/contact" class="dark:text-white">Contact</NavLi>
  </NavUl>
</Navbar>

<div class="page-background">
  <div class="chart-grid">
    {#each chartConfigs as config}
      <div class="chart-box">
        <h3 class="chart-title text-white">{config.label}</h3>
        <canvas id={config.id}></canvas>
      </div>
    {/each}
  </div>
</div>

<style type="postcss">
   :global(html) {
        background-image: linear-gradient(
            0deg,
            rgba(32, 42, 68, 1) 25%,
            rgba(9, 22, 46, 1) 50%,
            rgba(0, 1, 25, 1) 75%,
            rgba(0, 0, 0, 1) 100%
        );
        animation: slide 3s ease-in-out infinite alternate;
        background-attachment: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        z-index: -1;
        margin: auto auto;

    }
    .page-background{
      display: flex;
      justify-content: center;
    }

  @keyframes slide {
    0% {
      background-position: 0 0;
    }
    100% {
      background-position: 100% 100%;
    }
  }

  .chart-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
    width: 80%;
    max-width: 1000px;
  }

  .chart-box {
    background-color: #1f2937;
    border-radius: 0.5rem;
    padding: 1rem; /* Reduced padding */
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 100%; /* Make sure cards adjust within the grid */
    height: 100%; /* Adjust card height */
  }

  canvas {
    margin: auto;
  }

  .chart-title {
    text-align: center;
    margin-bottom: 1rem;
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .chart-grid {
      grid-template-columns: 1fr;
      margin-top: 5%;
    }
    :global(body) {
      overflow-y: scroll !important;
    }
  }
</style>
