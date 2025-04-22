<template>
    <q-item class="q-py-none">
        <q-item-section side class="text-weight-bolder"> ccTLDs </q-item-section>
        <q-item-section>
            <div class="flex justify-left q-gutter-sm q-pt-md">
                <q-checkbox :disable="disable" v-model="selectedTlds" :val="tld" :label="tld" color="cyan" v-for="tld in tldOptions" :key="tld" />
            </div>
            <div class="flex justify-left q-pt-xs q-pb-md q-px-sm">
                <q-input dense outlined :disable="disable" v-model="inputTld" placeholder="添加更多TLD">
                    <template v-slot:after>
                        <q-btn icon="add" color="primary" :disable="disable" @click="addTld" />
                    </template>
                </q-input>
            </div>
        </q-item-section>
    </q-item>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useQuasar } from "quasar";
import { useSettingStore } from "src/stores/settingStore";

defineOptions({
    name: "CcTldSelection"
});

const props = defineProps({
    disable: {
        type: Boolean,
        default: false
    }
});

const selectedTlds = defineModel("selectedTlds");

const $q = useQuasar();
const settingStore = useSettingStore();

const tldOptions = ref([]);
const inputTld = ref("");

function addTld() {
    if (inputTld.value) {
        let newTld = inputTld.value.replace(/\s+/g, "").replace(/^\.+/, "").replace(/\.+$/, "").toLowerCase();
        if (newTld.length > 0) {
            if (tldOptions.value.includes(newTld)) {
                $q.notify({
                    position: "top",
                    type: "warning",
                    message: "TLD已存在"
                });
            } else {
                tldOptions.value.push(newTld);
                selectedTlds.value.push(newTld);
                inputTld.value = "";
            }
        } else {
            $q.notify({
                position: "top",
                type: "warning",
                message: "输入的TLD无效"
            });
        }
    }
}

onMounted(() => {
    settingStore.typoDefaultCcTlds.forEach((tld) => {
        if (tld.isSelected) {
            selectedTlds.value.push(tld.tld);
        }
        tldOptions.value.push(tld.tld);
    });
});
</script>
