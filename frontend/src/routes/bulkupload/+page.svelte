<script lang="ts">
    import {
        Navbar,
        NavBrand,
        NavLi,
        NavUl,
        NavHamburger,
    } from "flowbite-svelte";
    import { Spinner } from "flowbite-svelte";
    import { onMount } from "svelte";
    import { EventsEmit, EventsOn } from "$lib/wailsjs/runtime/runtime";

    let spinnerVisible: boolean = false;
    let fileInput: HTMLInputElement | null = null;
    let progressMessage: string = "";
    type ProgressData = {
        current: number;
        total: number;
    };

    onMount(() => {
        const unsubscribeStarted = EventsOn(
            "hashingStarted",
            (total: number) => {
                alert(`Hashing started for ${total} passwords!`);
            },
        );

        const unsubscribeProgress = EventsOn(
            "hashingProgress",
            (data: ProgressData) => {
                progressMessage = `Processed ${data.current}/${data.total} passwords.`;
                console.log(progressMessage);
            },
        );

        const unsubscribeComplete = EventsOn("hashingCompleted", () => {
            spinnerVisible = false;
            alert("Hashing completed successfully!");
        });

        return () => {
            unsubscribeStarted();
            unsubscribeProgress();
            unsubscribeComplete();
        };
    });

    const handleFileUpload = async (): Promise<void> => {
        if (!fileInput?.files?.length) {
            alert("Please select a .txt file!");
            return;
        }

        const file: File = fileInput.files[0];
        if (file.type !== "text/plain") {
            alert("Only .txt files are allowed!");
            return;
        }

        spinnerVisible = true;
        progressMessage = "";

        try {
            const text = await file.text();
            const passwords = text
                .split("\n")
                .map((line) => line.trim())
                .filter((line) => line);

            if (passwords.length === 0) {
                alert("The file is empty or improperly formatted.");
                return;
            }

            console.log("Sending passwords to backend:", passwords);

            EventsEmit("StartHashing", passwords);
        } catch (error) {
            console.error("Error processing file:", error);
            alert("An error occurred while processing the file.");
        } finally {
            spinnerVisible = false;
        }
    };
</script>

<div>
    <Navbar rounded class="bg-transparent dark">
        <NavBrand href="/">
            <span class="self-center text-xl font-semibold dark:text-white"
                >Raingo</span
            >
        </NavBrand>
        <NavHamburger />
        <NavUl>
            <NavLi href="/" activeClass="active" class="dark:text-white"
                >Home</NavLi
            >
            <NavLi href="/checkhash" class="dark:text-white">Converter</NavLi>
            <NavLi href="/chart" class="dark:text-white">Stats</NavLi>
            <NavLi href="/bulkupload" class="dark:text-white">Upload</NavLi>
            <NavLi href="/table" class="dark:text-white">Table</NavLi>
        </NavUl>
    </Navbar>
</div>
<div class="card text-white">
    <div class="card-content">
        <h5 class="card-title font-extrabold">Bulk Upload Password</h5>
        <small>type: .txt | format: each password on each line</small>
        <div class="flex flex-col space-y-6">
            <input
                bind:this={fileInput}
                type="file"
                class="mt-5 px-4 py-3 rounded-lg bg-gray-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg file:mr-4 file:py-3 file:px-6 file:rounded-full file:border-0 file:text-lg file:font-semibold file:bg-violet-50 file:text-violet-700 hover:file:bg-violet-100"
            />
            <button
                class="bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-6 rounded-lg text-lg"
                on:click={handleFileUpload}
            >
                {#if spinnerVisible}
                    <Spinner size="sm" class="text-white" />
                {:else}
                    Upload
                {/if}
            </button>
        </div>
        {#if progressMessage}
            <p class="mt-4 text-sm text-gray-400">{progressMessage}</p>
        {/if}
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

    @media (max-width: 768px) {
        .card {
            width: 90%;
            height: auto;
            padding: 1.5rem;
            margin-top: 2rem;
        }
    }

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
