<template>
    <q-item class="q-py-none">
        <q-item-section side class="text-weight-bolder"> 注册接口 </q-item-section>
        <q-item-section>
            <div class="flex justify-left q-gutter-sm q-py-md">
                <q-radio
                    v-model="registerType"
                    checked-icon="check_circle"
                    unchecked-icon="radio_button_unchecked"
                    :val="api"
                    :label="api"
                    :disable="disable"
                    v-for="api in settingStore.registerApis"
                    :key="api"
                />
            </div>
        </q-item-section>
    </q-item>
</template>

<script setup>
import { useSettingStore } from "src/stores/settingStore";

defineOptions({
    name: "RegisterSelection"
});

const props = defineProps({
    disable: {
        type: Boolean,
        default: false
    }
});

const registerType = defineModel("registerType");

const settingStore = useSettingStore();
</script>
