import { defineStore } from "pinia";
import { date } from "quasar";

export const useWebStore = defineStore("web", {
    state: () => ({
        domains: [],
        unCheckDomains: 0,
        unRegisterDomains: 0,
        checkedTakenDomains: 0,
        checkedFreeDomains: 0,
        checkedErrorDomains: 0,
        isErrorRecheck: false
    }),

    actions: {
        setIsErrorRecheck(isErrorRecheck) {
            this.isErrorRecheck = isErrorRecheck;
        },
        setCheckingDomains(newDomains, checkType, typoType) {
            if (this.isErrorRecheck) {
                newDomains.forEach((domain) => {
                    for (let i = 0; i < this.domains.length; i++) {
                        if (this.domains[i].domain == domain.toLowerCase()) {
                            this.domains[i].lookupType = null;
                            this.domains[i].status = null;
                            this.domains[i].errorInfo = "";
                            this.domains[i].createdDate = null;
                            this.domains[i].expiryDate = null;
                            this.domains[i].nameServer = null;
                            this.domains[i].domainStatus = null;
                            this.domains[i].checkedRawResponse = null;
                            this.domains[i].isChecked = false;
                            this.domains[i].selectedRegister = false;
                            this.domains[i].registerStatus = null;
                            this.domains[i].registerRawResponse = null;
                            this.unCheckDomains++;
                            this.checkedErrorDomains--;
                        }
                    }
                });
            } else {
                let domainID = this.domains.length + 1;
                newDomains.forEach((domain) => {
                    this.domains.push({
                        domainID: domainID,
                        domain: domain.toLowerCase(),
                        checkType: checkType,
                        typoType: typoType,
                        lookupType: null,
                        status: null,
                        errorInfo: "",
                        createdDate: null,
                        expiryDate: null,
                        nameServer: null,
                        domainStatus: null,
                        checkedRawResponse: null,
                        isChecked: false,
                        selectedRegister: false,
                        registerStatus: null,
                        registerRawResponse: null
                    });
                    domainID++;
                    this.unCheckDomains++;
                });
            }
        },
        updateCheckResult(domainResult) {
            for (let i = 0; i < this.domains.length; i++) {
                if (this.domains[i].domain == domainResult.domain.toLowerCase()) {
                    this.domains[i].status = domainResult.registerStatus;
                    this.domains[i].lookupType = domainResult.lookupType;
                    this.domains[i].checkedRawResponse = domainResult.rawResponse;

                    if (domainResult.registerStatus == "Taken") {
                        if (domainResult.lookupType == "whois" || domainResult.lookupType == "rdap") {
                            this.domains[i].createdDate = date.formatDate(domainResult.createdDate, "YYYY-MM-DD");
                            this.domains[i].expiryDate = date.formatDate(domainResult.expiryDate, "YYYY-MM-DD");
                            this.domains[i].domainStatus = domainResult.domainStatus;
                        }

                        if (domainResult.nameServer && domainResult.nameServer.length > 0) {
                            this.domains[i].nameServer = domainResult.nameServer.map((item) => item.toLowerCase());
                        } else {
                            this.domains[i].nameServer = ["N/A"];
                        }

                        this.checkedTakenDomains++;
                    } else if (domainResult.registerStatus == "Free") {
                        this.checkedFreeDomains++;
                    } else if (domainResult.registerStatus == "Error") {
                        this.domains[i].errorInfo = domainResult.queryError;
                        this.checkedErrorDomains++;
                    }

                    this.domains[i].isChecked = true;
                    this.unCheckDomains--;
                }
            }
        },
        setRegisteringDomains(domains) {
            domains.forEach((domain) => {
                for (let i = 0; i < this.domains.length; i++) {
                    if (this.domains[i].domainID == domain.domainID) {
                        this.domains[i].selectedRegister = true;
                        this.domains[i].registerStatus = null;
                        this.domains[i].registerRawResponse = null;
                        this.unRegisterDomains++;
                    }
                }
            });
        },
        updateRegisterResult(registerResult) {
            for (let i = 0; i < this.domains.length; i++) {
                if (this.domains[i].domain == registerResult.domainName && this.domains[i].selectedRegister) {
                    if (this.domains[i].registerStatus == null) {
                        this.domains[i].registerStatus = registerResult.registerStatus;
                        this.domains[i].registerRawResponse = registerResult.rawResponse;
                        this.unRegisterDomains--;
                    }
                }
            }
        },
        clearDomains() {
            this.domains = [];
            this.unCheckDomains = 0;
            this.unRegisterDomains = 0;
            this.checkedTakenDomains = 0;
            this.checkedFreeDomains = 0;
            this.checkedErrorDomains = 0;
            this.isErrorRecheck = false;
        }
    }
});
