export namespace engine {
	
	export class JobReport {
	    InputName: string;
	    OutputDir: string;
	    Status: string;
	    SourceSize: number;
	    TargetSize: number;
	    Duration: number;
	    ErrMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new JobReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InputName = source["InputName"];
	        this.OutputDir = source["OutputDir"];
	        this.Status = source["Status"];
	        this.SourceSize = source["SourceSize"];
	        this.TargetSize = source["TargetSize"];
	        this.Duration = source["Duration"];
	        this.ErrMessage = source["ErrMessage"];
	    }
	}
	export class JobRequest {
	    Files: string[];
	    Preset: string;
	    HWAccel: boolean;
	    CopyAudio: boolean;
	    ForceMode: boolean;
	    SkipExisting: boolean;
	    Concurrency: number;
	
	    static createFrom(source: any = {}) {
	        return new JobRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Files = source["Files"];
	        this.Preset = source["Preset"];
	        this.HWAccel = source["HWAccel"];
	        this.CopyAudio = source["CopyAudio"];
	        this.ForceMode = source["ForceMode"];
	        this.SkipExisting = source["SkipExisting"];
	        this.Concurrency = source["Concurrency"];
	    }
	}

}

export namespace main {
	
	export class PresetInfo {
	    name: string;
	    scaleFactor: number;
	    maxDimension: number;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new PresetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.scaleFactor = source["scaleFactor"];
	        this.maxDimension = source["maxDimension"];
	        this.description = source["description"];
	    }
	}

}

