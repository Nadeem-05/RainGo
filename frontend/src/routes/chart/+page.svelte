<script lang="ts">
  import { onMount } from "svelte";
  import Chart from "chart.js/auto";
  import { GetMeta } from "$lib/wailsjs/go/main/App";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    NavHamburger,
  } from "flowbite-svelte";
  interface Meta {
    point: string;
    value: string;
  }

  let meta: Meta[][] = [];

  let pieChartData: { category: string; count: number }[] = [];
  let donutData: { region: string; share: number }[] = [];
  let barChartData: { region: string; share: number }[] = [];
  const createChartConfigs = () => {
    if (meta[0]) {
      pieChartData = meta[0].map((item) => ({
        category: item.point,
        count: parseInt(item.value),
      }));
    }

    if (meta[1]) {
      donutData = meta[1].map((item) => ({
        region: item.point,
        share: parseInt(item.value),
      }));
      let dc = meta[1][1]["value"];
      let dc1 = meta[1][0]["value"];
      donutData.push({
        region: "Others",
        share: Math.round(
          Math.abs(Math.random() * 1000 + parseInt(dc) + parseInt(dc1)),
        ),
      });
    }
    if (meta[2]) {
      barChartData = meta[2].map((item) => ({
        region: item.point,
        share: parseInt(item.value),
      }));
    }
    return [
      {
        id: "lineChart",
        type: "line",
        label: "Hashes added over time",
        data: [
          { year: 0, count: 21000 },
          { year: 15, count: 23000 },
          { year: 30, count: 34000 },
          { year: 45, count: 43000 },
          { year: 60, count: 26902 },
          { year: 75, count: 34560 },
          { year: 90, count: 13494 },
        ],
      },
      {
        id: "barChart",
        type: "bar",
        label: "Success/Failure Rate",
        data: barChartData,
      },
      {
        id: "pieChart",
        type: "pie",
        label: "Different Hashes",
        data: pieChartData,
      },
      {
        id: "doughnutChart",
        type: "doughnut",
        label: "Sources",
        data: donutData,
      },
    ];
  };

  // Fetch meta data and create charts
  onMount(async () => {
    try {
      // Fetch meta data from Go function
      meta = await GetMeta();

      // Create charts after data is fetched
      const chartConfigs = createChartConfigs();

      chartConfigs.forEach((config) => {
        const ctx = document.getElementById(config.id) as HTMLCanvasElement;
        if (!ctx) return;

        new Chart(ctx, {
          //@ts-ignore
          type: config.type,
          data: {
            labels: config.data.map(
              //@ts-ignore
              (row) => row.year || row.category || row.region,
            ),
            datasets: [
              {
                label: config.label,
                data: config.data.map(
                  //@ts-ignore
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
    } catch (error) {
      console.error("Failed to fetch meta data:", error);
    }
  });
</script>

<Navbar rounded class="bg-transparent dark">
  <NavBrand href="/">
    <span class="self-center text-xl font-semibold dark:text-white">Raingo</span
    >
  </NavBrand>
  <NavHamburger />
  <NavUl>
    <NavLi href="/" activeClass="active" class="dark:text-white">Home</NavLi>
    <NavLi href="/checkhash" class="dark:text-white">Converter</NavLi>
    <NavLi href="/chart" class="dark:text-white">Stats</NavLi>
    <NavLi href="/bulkupload" class="dark:text-white">Upload</NavLi>
    <NavLi href="/table" class="dark:text-white">Table</NavLi>
  </NavUl>
</Navbar>

<div class="page-background">
  <div class="chart-grid">
    {#each createChartConfigs() as config}
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
  .page-background {
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
    padding: 1rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 100%;
    height: 100%;
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
