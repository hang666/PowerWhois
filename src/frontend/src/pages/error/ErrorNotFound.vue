<template>
    <q-layout>
        <q-page-container>
            <q-page class="flex flex-center">
                <div class="text-center">
                    <img src="/images/404.svg" alt="" />
                    <div class="text-center">
                        <h5 class="text-grey text-center">抱歉，你访问的页面不存在</h5>
                        <q-btn push color="primary" label="返回首页" to="/" />
                    </div>
                </div>
            </q-page>
        </q-page-container>
    </q-layout>
</template>

<script setup>
defineOptions({
    name: "ErrorNotFound"
});
</script>

<style>
img {
    width: 350px;
    margin: 0 auto;
}
</style>
