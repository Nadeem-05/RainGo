<script>
    import {
        Navbar,
        NavBrand,
        NavLi,
        NavUl,
        NavHamburger,
    } from "flowbite-svelte";
    import { Spinner } from "flowbite-svelte";
    import { GetPassword } from "$lib/wailsjs/go/main/App";
    import { onMount } from "svelte";
    //@ts-ignore
    let inputElement;
    let password = "";
    let spinnerVisible = false;
    let textArea = false;
    // @ts-ignore
    const sleep = (seconds) =>
        new Promise((resolve) => setTimeout(resolve, seconds * 1000));

    function handleClick() {
        textArea = false;
            //@ts-ignore
        if (!inputElement.value) {
            alert("Please enter a hash");
            return;
        }
        spinnerVisible = true;
        sleep(1).then(() => {
            spinnerVisible = false;
                //@ts-ignore
            GetPassword(inputElement.value).then((res) => {
                if (res === "Failed") {
                    alert("Hash not found");
                } else {
                    textArea = true;
                    password = res;
                }
            });
        });
    }

    onMount(() => {});
</script>
<div>
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
        <NavLi href="/docs/components/navbar" class="dark:text-white"
          >Navbar</NavLi
        >
        <NavLi href="/pricing" class="dark:text-white">Pricing</NavLi>
        <NavLi href="/contact" class="dark:text-white">Contact</NavLi>
      </NavUl>
    </Navbar>
  </div>

<div class="card text-white">
    <div class="card-content">
        <h5 class="card-title font-extrabold">Try with a hash</h5>
        <small>eg: hash</small>
        <div class="flex flex-col space-y-6">
            <input
                bind:this={inputElement}
                type="text"
                class="mt-5 px-4 py-3 rounded-lg bg-gray-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg"
                placeholder="Enter hash below"
            />
            <button
                on:click={handleClick}
                class="bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-6 rounded-lg text-lg"
            >
                Submit
            </button>
            {#if spinnerVisible}
                <div class="flex justify-center">
                    <Spinner size="8" class="text-blue-500" />
                </div>
            {/if}
            {#if textArea}
            <div>
                <textarea 
                    class="mt-2 px-4 py-1 rounded-lg bg-gray-800 h-24 w-192 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg"
                    rows="5"
                    placeholder="Password will be displayed here"
                    readonly>{password}</textarea
                >
            </div>
            {/if}
        </div>
    </div>
</div>

<style>
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

    @keyframes slide {
        0% {
            background-position: 0 0;
        }
        100% {
            background-position: 100% 100%;
        }
    }

    .card {
        background-color: #1f2937;
        border-radius: 0.5rem;
        padding: 2rem;
        box-shadow:
            0 4px 6px -1px rgba(0, 0, 0, 0.1),
            0 2px 4px -1px rgba(0, 0, 0, 0.06);
        transition: box-shadow 0.3s ease;
        position: relative;
        width: 50%;
        height: 30rem;
        margin: auto auto;
        text-align: center;
    }

    /* Medium screens (tablets) */
    @media (max-width: 1024px) {
        .card {
            width: 70%;
            height: 25rem;
            margin-top: 5rem;
        }
    }

    /* Small screens (mobile) */
    @media (max-width: 768px) {
        .card {
            width: 90%;
            height: auto;
            padding: 1.5rem;
            margin-top: 2rem;
        }
    }

    /* Extra-small screens (small mobile) */
    @media (max-width: 480px) {
        .card {
            width: 95%;
            padding: 1rem;
        }
    }

    .card:hover {
        box-shadow:
            0 10px 15px -3px rgba(0, 0, 0, 0.1),
            0 4px 6px -2px rgba(0, 0, 0, 0.05);
    }
</style>
