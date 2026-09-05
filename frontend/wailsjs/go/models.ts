export namespace engine {
	
	export class JobReport {
	    InputName: string;
	    InputPath: string;
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
	        this.InputPath = source["InputPath"];
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
	    OutputDir: string;
	    VideoCodec: string;
	    MaxFPS: number;
	    AudioMode: string;
	    CRF: number;
	
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
	        this.OutputDir = source["OutputDir"];
	        this.VideoCodec = source["VideoCodec"];
	        this.MaxFPS = source["MaxFPS"];
	        this.AudioMode = source["AudioMode"];
	        this.CRF = source["CRF"];
	    }
	}

}

export namespace gif {
	
	export class ExportRequest {
	    Files: string[];
	    Tier: string;
	    Formats: string[];
	    OutputDir: string;
	    Force: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Files = source["Files"];
	        this.Tier = source["Tier"];
	        this.Formats = source["Formats"];
	        this.OutputDir = source["OutputDir"];
	        this.Force = source["Force"];
	    }
	}
	export class ExportResult {
	    inputName: string;
	    inputPath: string;
	    outputPath: string;
	    format: string;
	    tier: string;
	    size: number;
	    status: string;
	    error?: string;
	    overBudget: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputName = source["inputName"];
	        this.inputPath = source["inputPath"];
	        this.outputPath = source["outputPath"];
	        this.format = source["format"];
	        this.tier = source["tier"];
	        this.size = source["size"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.overBudget = source["overBudget"];
	    }
	}

}

export namespace gui {
	
	export class GifTierInfo {
	    name: string;
	    maxWidth: number;
	    maxSizeMB: number;
	    fps: number;
	    maxDuration: string;
	    description: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GifTierInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.maxWidth = source["maxWidth"];
	        this.maxSizeMB = source["maxSizeMB"];
	        this.fps = source["fps"];
	        this.maxDuration = source["maxDuration"];
	        this.description = source["description"];
	        this.isDefault = source["isDefault"];
	    }
	}
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

