<template>
    <q-item class="q-py-none">
        <q-item-section side class="text-weight-bolder"> Typo类型 </q-item-section>
        <q-item-section>
            <div class="flex justify-left q-gutter-sm q-py-md">
                <q-checkbox
                    :disable="disable"
                    v-model="typoTypeSelected"
                    :val="typoType"
                    :label="typoLabel"
                    color="cyan"
                    v-for="(typoLabel, typoType) in typoTypes"
                    :key="typoType"
                />
            </div>
        </q-item-section>
    </q-item>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { typoTypes } from "src/utils/globalDefinition";

defineOptions({
    name: "TypoSelection"
});

const props = defineProps({
    disable: {
        type: Boolean,
        default: false
    }
});

const typoTypeSelected = defineModel("typoTypeSelected", {
    default: () => ["www", "skipLetter", "wrongHorizontalKey"]
});

onMounted(() => {
    // 确保在组件挂载时有默认选中的值
    if (typoTypeSelected.value.length === 0) {
        typoTypeSelected.value = ["www", "skipLetter", "wrongHorizontalKey"];
    }
});
</script>
